package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yklcs/agent-hub-mcp/internal/config"
	"github.com/yklcs/agent-hub-mcp/internal/db"
)

// runDoctor performs system diagnostics.
func (a *App) runDoctor(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	dbPath := fs.String("db", config.DefaultDBPath(), "Path to SQLite database")
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
