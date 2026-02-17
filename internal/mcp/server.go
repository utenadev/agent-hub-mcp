package mcp

import (
	"log"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/yklcs/agent-hub-mcp/internal/db"
)

// Server wraps the MCP server with our database.
type Server struct {
	mcpServer     *server.MCPServer
	db            *db.DB
	DefaultSender string
}

// NewServer creates a new MCP server with the given database and default sender.
func NewServer(database *db.DB, defaultSender string) *Server {
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"agent-hub-mcp",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	s := &Server{
		mcpServer:     mcpServer,
		db:            database,
		DefaultSender: defaultSender,
	}

	// Register tools
	s.registerTools()

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
