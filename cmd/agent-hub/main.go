package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/yklcs/agent-hub-mcp/internal/db"
	"github.com/yklcs/agent-hub-mcp/internal/hub"
	"github.com/yklcs/agent-hub-mcp/internal/mcp"
)

// App holds dependencies for testing
type App struct {
	Logger   *log.Logger
	ExitFunc func(int)
}

// NewApp creates a new App with defaults
func NewApp() *App {
	return &App{
		Logger:   log.New(os.Stderr, "", log.LstdFlags),
		ExitFunc: os.Exit,
	}
}

func main() {
	app := NewApp()
	if err := app.Run(os.Args, os.Stdin, os.Stdout, os.Stderr); err != nil {
		app.Logger.Printf("Error: %v\n", err)
		app.ExitFunc(1)
	}
}

// Run executes the application with given arguments and IO
func (a *App) Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) < 2 {
		a.runHelp(stdout)
		return nil
	}

	command := args[1]

	switch command {
	case "serve":
		return a.runServe(args[2:], stdout, stderr)
	case "orchestrator":
		return a.runOrchestrator(args[2:], stdout, stderr)
	case "doctor":
		return a.runDoctor(args[2:], stdout, stderr)
	case "setup":
		return a.runSetup(args[2:], stdout, stderr)
	case "help", "--help", "-h":
		a.runHelp(stdout)
		return nil
	default:
		return fmt.Errorf("Unknown command: %s\nRun 'agent-hub help' for usage", command)
	}
}

// getDefaultDBPath returns the standard database path.
func getDefaultDBPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "agent-hub.db"
	}
	return filepath.Join(configDir, "agent-hub-mcp", "agent-hub.db")
}

// runServe starts the MCP server.
func (a *App) runServe(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Suppress flag errors during testing
	dbPath := fs.String("db", getDefaultDBPath(), "Path to SQLite database")
	sseAddr := fs.String("sse", "", "Enable SSE mode on address (e.g., :8080)")
	senderFlag := fs.String("sender", "", "Default sender name for messages (overrides BBS_AGENT_ID env var)")
	roleFlag := fs.String("role", "", "Agent role (overrides BBS_AGENT_ROLE env var)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Determine sender: flag > env var > default ("unknown")
	sender := *senderFlag
	if sender == "" {
		sender = os.Getenv("BBS_AGENT_ID")
		if sender == "" {
			sender = "unknown"
		}
	}

	// Determine role: flag > env var > default ("agent")
	role := *roleFlag
	if role == "" {
		role = os.Getenv("BBS_AGENT_ROLE")
		if role == "" {
			role = "agent"
		}
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Check database integrity before starting
	if results, err := database.CheckIntegrity(); err != nil {
		// Print detailed error information to stderr
		fmt.Fprintf(stderr, "Database check failed:\n")
		for table, exists := range results {
			status := "OK"
			if !exists {
				status = "MISSING"
			}
			fmt.Fprintf(stderr, "  - %s: %s\n", table, status)
		}
		return fmt.Errorf("database integrity check failed: %w (run 'agent-hub setup' if this is a new installation)", err)
	}

	// Register/update agent presence
	if err := database.UpsertAgentPresence(sender, role); err != nil {
		fmt.Fprintf(stderr, "Warning: failed to register agent presence: %v\n", err)
	}

	fmt.Fprintf(stderr, "Database opened: %s\n", *dbPath)
	fmt.Fprintf(stderr, "Agent: name=%s, role=%s\n", sender, role)

	srv := mcp.NewServer(database, sender, role)

	if *sseAddr != "" {
		host := *sseAddr
		if host[0] == ':' {
			host = "localhost" + host
		}
		fmt.Fprintf(stderr, "Starting MCP server on SSE http://%s...\n", *sseAddr)
		fmt.Fprintf(stderr, "\n--- SSE Connection Info ---\n")
		fmt.Fprintf(stderr, "SSE Endpoint:     http://%s/sse\n", host)
		fmt.Fprintf(stderr, "Message Endpoint: http://%s/message\n", host)
		fmt.Fprintf(stderr, "---------------------------\n\n")
		if err := srv.ServeSSE(*sseAddr); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	} else {
		fmt.Fprintln(stderr, "Starting MCP server on stdio...")
		if err := srv.Serve(); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	return nil
}

// runOrchestrator starts the orchestrator monitor.
func (a *App) runOrchestrator(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("orchestrator", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dbPath := fs.String("db", getDefaultDBPath(), "Path to SQLite database")
	senderFlag := fs.String("sender", "", "Default sender name for messages (overrides BBS_AGENT_ID env var)")
	roleFlag := fs.String("role", "", "Agent role (overrides BBS_AGENT_ROLE env var)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Determine sender: flag > env var > default ("orchestrator")
	sender := *senderFlag
	if sender == "" {
		sender = os.Getenv("BBS_AGENT_ID")
		if sender == "" {
			sender = "orchestrator"
		}
	}

	// Determine role: flag > env var > default ("orchestrator")
	role := *roleFlag
	if role == "" {
		role = os.Getenv("BBS_AGENT_ROLE")
		if role == "" {
			role = "orchestrator"
		}
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Register/update agent presence
	if err := database.UpsertAgentPresence(sender, role); err != nil {
		fmt.Fprintf(stderr, "Warning: failed to register agent presence: %v\n", err)
	}

	fmt.Fprintf(stderr, "Orchestrator started with database: %s (agent: %s, role: %s)\n", *dbPath, sender, role)

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(stderr, "\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	orchestrator := hub.NewOrchestrator(database, hub.DefaultConfig())
	if err := orchestrator.Start(ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("orchestrator error: %w", err)
	}

	fmt.Fprintln(stderr, "Orchestrator stopped")
	return nil
}

// runDoctor performs system diagnostics.
func (a *App) runDoctor(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	dbPath := fs.String("db", getDefaultDBPath(), "Path to SQLite database")
	if err := fs.Parse(args); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "--- Agent Hub Doctor ---")
	allOk := true

	// Check Database
	fmt.Fprintf(stdout, "[*] Checking Database (%s)...\n", *dbPath)
	database, err := db.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(stdout, "  [ERROR] Failed to open database: %v\n", err)
		allOk = false
	} else {
		defer database.Close()
		results, err := database.CheckIntegrity()
		if err != nil {
			fmt.Fprintf(stdout, "  [ERROR] Integrity check failed: %v\n", err)
			fmt.Fprintln(stdout, "  Table status:")
			for table, exists := range results {
				status := "OK"
				if !exists {
					status = "MISSING"
				}
				fmt.Fprintf(stdout, "    - %s: %s\n", table, status)
			}
			allOk = false
		} else {
			fmt.Fprintln(stdout, "  [OK] Database integrity check passed")
			fmt.Fprintln(stdout, "  Table status:")
			// Pre-defined table order for consistent output
			tables := []string{"topics", "messages", "topic_summaries", "agent_presence"}
			for _, table := range tables {
				fmt.Fprintf(stdout, "    - %s: OK\n", table)
			}
		}
	}

	// Check Database File Permissions
	fmt.Fprintf(stdout, "[*] Checking Database Permissions (%s)... ", *dbPath)
	if _, err := os.Stat(*dbPath); err == nil {
		// Check read permission
		file, err := os.Open(*dbPath)
		if err != nil {
			fmt.Fprintf(stdout, "[ERROR] Cannot read database file: %v\n", err)
			allOk = false
		} else {
			file.Close()
			// Check write permission by attempting to open for append
			file, err = os.OpenFile(*dbPath, os.O_WRONLY, 0)
			if err != nil {
				fmt.Fprintf(stdout, "[ERROR] Cannot write to database file: %v\n", err)
				allOk = false
			} else {
				file.Close()
				fmt.Fprintln(stdout, "[OK]")
			}
		}
	} else if os.IsNotExist(err) {
		// Database doesn't exist yet - check parent directory permissions
		dbDir := filepath.Dir(*dbPath)
		if dbDir == "." {
			dbDir, _ = os.Getwd()
		}
		info, err := os.Stat(dbDir)
		if err != nil {
			fmt.Fprintf(stdout, "[WARN] Database directory not accessible: %v\n", err)
		} else {
			fmt.Fprintf(stdout, "[OK] Database will be created in: %s\n", dbDir)
			_ = info
		}
	} else {
		fmt.Fprintf(stdout, "[ERROR] Cannot access database path: %v\n", err)
		allOk = false
	}

	// Check Config Directory
	configDir, err := os.UserConfigDir()
	if err == nil {
		agentHubDir := filepath.Join(configDir, "agent-hub-mcp")
		fmt.Fprintf(stdout, "[*] Checking Config Directory (%s)... ", agentHubDir)
		info, err := os.Stat(agentHubDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintln(stdout, "[WARN] Directory does not exist. Run 'agent-hub setup' to create it.")
			} else {
				fmt.Fprintf(stdout, "[ERROR] Cannot access directory: %v\n", err)
				allOk = false
			}
		} else {
			if info.Mode().Perm()&0600 == 0600 {
				fmt.Fprintln(stdout, "[OK]")
			} else {
				fmt.Fprintf(stdout, "[WARN] Directory permissions are %v, may cause issues\n", info.Mode().Perm())
			}
		}
	}

	// Check API Keys
	fmt.Fprint(stdout, "[*] Checking Gemini API Key... ")
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("HUB_MASTER_API_KEY")
	}
	if apiKey == "" {
		fmt.Fprintln(stdout, "[WARN] Neither GEMINI_API_KEY nor HUB_MASTER_API_KEY is set. Summarization will be mocked.")
	} else {
		fmt.Fprintf(stdout, "[OK] (ends with %s)\n", apiKey[len(apiKey)-4:])
	}

	// Check Agent Config
	fmt.Fprint(stdout, "[*] Checking Agent Configuration... ")
	sender := os.Getenv("BBS_AGENT_ID")
	role := os.Getenv("BBS_AGENT_ROLE")
	configOk := true
	if sender == "" {
		fmt.Fprint(stdout, "\n  [WARN] BBS_AGENT_ID not set. Will default to 'unknown'.")
		configOk = false
	}
	if role == "" {
		fmt.Fprint(stdout, "\n  [WARN] BBS_AGENT_ROLE not set. Will default to 'agent'.")
		configOk = false
	}
	if configOk {
		fmt.Fprintf(stdout, "[OK] (name=%s, role=%s)\n", sender, role)
	} else {
		fmt.Fprintln(stdout)
	}

	if allOk {
		fmt.Fprintln(stdout, "\nResult: All critical checks passed!")
	} else {
		fmt.Fprintln(stdout, "\nResult: Some issues found. Please check the errors above.")
		fmt.Fprintln(stdout, "Run 'agent-hub setup' if this is a fresh installation.")
		return fmt.Errorf("diagnostics failed")
	}
	return nil
}

// runSetup initializes the system.
func (a *App) runSetup(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	dbPath := fs.String("db", getDefaultDBPath(), "Path to SQLite database")
	force := fs.Bool("force", false, "Overwrites existing configuration if present")
	if err := fs.Parse(args); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "--- Agent Hub Setup ---")

	// 1. Initialize Database
	if _, err := os.Stat(*dbPath); err == nil && !*force {
		fmt.Fprintf(stdout, "[*] Database already exists at %s. Use -force to overwrite.\n", *dbPath)
	} else {
		fmt.Fprintf(stdout, "[*] Initializing Database at %s... ", *dbPath)
		database, err := db.Open(*dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		database.Close()
		fmt.Fprintln(stdout, "OK")
	}

	// 2. Create Config Directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	agentHubDir := filepath.Join(configDir, "agent-hub-mcp")
	fmt.Fprintf(stdout, "[*] Creating Config Directory (%s)... ", agentHubDir)
	if err := os.MkdirAll(agentHubDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	fmt.Fprintln(stdout, "OK")

	// 3. Create config.json template
	configPath := filepath.Join(agentHubDir, "config.json")
	if _, err := os.Stat(configPath); err == nil && !*force {
		fmt.Fprintf(stdout, "[*] Config file already exists at %s. Use -force to overwrite.\n", configPath)
	} else {
		fmt.Fprintf(stdout, "[*] Creating Config Template (%s)... ", configPath)
		config := map[string]string{
			"gemini_api_key": "YOUR_GEMINI_API_KEY_HERE",
			"default_sender": "",
			"default_role":   "",
		}
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		fmt.Fprintln(stdout, "OK")
	}

	// 4. Claude Desktop Config Helper
	fmt.Fprintln(stdout, "\n--- Claude Desktop Configuration ---")

	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		execPath = "agent-hub"
	}

	fmt.Fprintln(stdout, "Add the following to your Claude Desktop configuration:")
	fmt.Fprintln(stdout)

	// OS-specific config path
	var claudeConfigPath string
	switch runtime.GOOS {
	case "darwin":
		claudeConfigPath = "~/Library/Application Support/Claude/claude_desktop_config.json"
	case "windows":
		claudeConfigPath = "%APPDATA%/Claude/claude_desktop_config.json"
	default: // linux and others
		claudeConfigPath = "~/.config/Claude/claude_desktop_config.json"
	}

	fmt.Fprintf(stdout, "Config file location: %s\n", claudeConfigPath)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "JSON snippet to add:")
	fmt.Fprintln(stdout, "```json")
	fmt.Fprintf(stdout, "{\n  \"mcpServers\": {\n    \"agent-hub\": {\n      \"command\": \"%s\",\n      \"args\": [\"serve\", \"-db\", \"%s\"]\n    }\n  }\n}\n", execPath, *dbPath)
	fmt.Fprintln(stdout, "```")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Setup complete! You can now run 'agent-hub serve' or 'agent-hub doctor'.")

	return nil
}

// runHelp displays usage information.
func (a *App) runHelp(stdout io.Writer) {
	fmt.Fprintln(stdout, "Agent Hub MCP - Multi-Agent Collaboration BBS")
	fmt.Fprintln(stdout, "\nUsage:")
	fmt.Fprintln(stdout, "  agent-hub <command> [args...]")
	fmt.Fprintln(stdout, "\nCommands:")
	fmt.Fprintln(stdout, "  serve         Start the MCP server (stdio or SSE mode)")
	fmt.Fprintln(stdout, "  orchestrator  Start the autonomous monitor/summarizer")
	fmt.Fprintln(stdout, "  doctor        Run system diagnostics")
	fmt.Fprintln(stdout, "  setup         Initialize database and configuration")
	fmt.Fprintln(stdout, "  help          Show this help message")
	fmt.Fprintln(stdout, "\nGlobal Flags (available for most commands):")
	fmt.Fprintln(stdout, "  -db string    Path to SQLite database (default: "+getDefaultDBPath()+")")
	fmt.Fprintln(stdout, "\nServe Flags:")
	fmt.Fprintln(stdout, "  -sse string   Enable SSE mode on address (e.g., :8080)")
	fmt.Fprintln(stdout, "  -sender name  Default sender name for messages")
	fmt.Fprintln(stdout, "  -role role    Agent role")
	fmt.Fprintln(stdout, "\nSSE Connection Example:")
	fmt.Fprintln(stdout, "  When running with '-sse :8080', connect your MCP client to:")
	fmt.Fprintln(stdout, "  http://localhost:8080/sse")
}
