package mcp

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/yklcs/agent-hub-mcp/internal/db"
)

// Server wraps the MCP server with our database.
type Server struct {
	mcpServer     *server.MCPServer
	db            *db.DB
	DefaultSender string
	DefaultRole   string
	CurrentSender string
	notifier      *db.Notifier
}

// getSender returns the current sender, falling back to default if not set.
func (s *Server) getSender() string {
	if s.CurrentSender != "" {
		return s.CurrentSender
	}
	if s.DefaultSender != "" {
		return s.DefaultSender
	}
	return "unknown"
}

// NewServer creates a new MCP server with the given database, default sender, and role.
func NewServer(database *db.DB, defaultSender, defaultRole string) *Server {
	// Create MCP server with tool and resource capabilities
	mcpServer := server.NewMCPServer(
		"agent-hub-mcp",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	s := &Server{
		mcpServer:     mcpServer,
		db:            database,
		DefaultSender: defaultSender,
		DefaultRole:   defaultRole,
	}

	// Register tools
	s.registerTools()

	// Register resources
	s.registerResources()

	return s
}

// registerTools registers all BBS tools.
func (s *Server) registerTools() {
	// bbs_create_topic tool
	createTopicTool := mcp.NewTool(
		"bbs_create_topic",
		mcp.WithDescription("Create a new discussion topic"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("The title of the topic"),
		),
	)

	s.mcpServer.AddTool(createTopicTool, s.handleBBSCreateTopic)

	// bbs_post tool
	postTool := mcp.NewTool(
		"bbs_post",
		mcp.WithDescription("Post a message to a topic"),
		mcp.WithNumber("topic_id",
			mcp.Required(),
			mcp.Description("The ID of the topic"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The message content"),
		),
	)

	s.mcpServer.AddTool(postTool, s.handleBBSPost)

	// bbs_read tool
	readTool := mcp.NewTool(
		"bbs_read",
		mcp.WithDescription("Read recent messages from a topic"),
		mcp.WithNumber("topic_id",
			mcp.Required(),
			mcp.Description("The ID of the topic"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of messages to return (default: 10)"),
		),
	)

	s.mcpServer.AddTool(readTool, s.handleBBSRead)

	// check_hub_status tool
	checkHubStatusTool := mcp.NewTool(
		"check_hub_status",
		mcp.WithDescription("Check hub status for unread messages and team presence"),
	)

	s.mcpServer.AddTool(checkHubStatusTool, s.handleCheckHubStatus)

	// update_status tool
	updateStatusTool := mcp.NewTool(
		"update_status",
		mcp.WithDescription("Update your current status and working topic"),
		mcp.WithString("status",
			mcp.Required(),
			mcp.Description("Your current status (e.g., 'implementing', 'testing', 'waiting')"),
		),
		mcp.WithNumber("topic_id",
			mcp.Description("Current topic ID you are working on (optional)"),
		),
	)

	s.mcpServer.AddTool(updateStatusTool, s.handleUpdateStatus)

	// bbs_register_agent tool
	registerAgentTool := mcp.NewTool(
		"bbs_register_agent",
		mcp.WithDescription("Register or update your agent identity in the hub"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Your agent identifier"),
		),
		mcp.WithString("role",
			mcp.Required(),
			mcp.Description("Your agent role (e.g., coder, reviewer, architect)"),
		),
		mcp.WithString("status",
			mcp.Description("Initial status (default: online)"),
		),
		mcp.WithNumber("topic_id",
			mcp.Description("Current topic ID you're working on (optional)"),
		),
	)

	s.mcpServer.AddTool(registerAgentTool, s.handleRegisterAgent)

	// wait_notify tool
	waitNotifyTool := mcp.NewTool(
		"wait_notify",
		mcp.WithDescription("Wait for new messages on a topic (long-polling)"),
		mcp.WithString("agent_id",
			mcp.Required(),
			mcp.Description("The agent identifier waiting for notifications"),
		),
		mcp.WithNumber("timeout_sec",
			mcp.Description("Timeout in seconds (default: 180)"),
		),
	)

	s.mcpServer.AddTool(waitNotifyTool, s.handleWaitNotify)
}

// readGuidelines reads the agent collaboration guidelines from the docs directory.
func (s *Server) readGuidelines() string {
	content, err := os.ReadFile("docs/AGENTS_SYSTEM_PROMPT.md")
	if err != nil {
		log.Printf("Warning: could not read guidelines file: %v", err)
		return "# Guidelines\n\nGuidelines file not found."
	}
	return string(content)
}

// registerResources registers MCP resources for agent guidelines.
func (s *Server) registerResources() {
	guidelinesResource := mcp.NewResource(
		"guidelines://agent-collaboration",
		"Agent Collaboration Guidelines",
		mcp.WithResourceDescription("Guidelines for multi-agent collaboration via BBS"),
		mcp.WithMIMEType("text/markdown"),
	)

	s.mcpServer.AddResource(guidelinesResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		content := s.readGuidelines()
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "guidelines://agent-collaboration",
				MIMEType: "text/markdown",
				Text:     content,
			},
		}, nil
	})
}

// Serve starts the MCP server on stdio.
func (s *Server) Serve() error {
	log.Println("Starting MCP server on stdio...")
	return server.ServeStdio(s.mcpServer)
}

// ServeSSE starts the MCP server on an HTTP endpoint with SSE.
func (s *Server) ServeSSE(addr string) error {
	log.Printf("Starting MCP server on SSE http://%s...", addr)

	sseServer := server.NewSSEServer(s.mcpServer,
		server.WithBasePath("/"),
	)

	mux := http.NewServeMux()
	mux.Handle("/", sseServer)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("SSE server listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
