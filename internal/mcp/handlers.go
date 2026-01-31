package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleBBSCreateTopic handles the bbs_create_topic tool.
func (s *Server) handleBBSCreateTopic(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, err := req.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError("title is required and must be a string"), nil
	}

	id, err := s.db.CreateTopic(title)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create topic: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Topic created with ID: %d", id)), nil
}

// handleBBSPost handles the bbs_post tool.
func (s *Server) handleBBSPost(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	topicID, err := req.RequireFloat("topic_id")
	if err != nil {
		return mcp.NewToolResultError("topic_id is required and must be a number"), nil
	}

	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required and must be a string"), nil
	}

	// Get sender from context or default to "unknown"
	// TODO: Extract sender from BBS_AGENT_ID environment variable or request context
	sender := "unknown"

	id, err := s.db.PostMessage(int64(topicID), sender, content)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to post message: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Message posted with ID: %d", id)), nil
}

// handleBBSRead handles the bbs_read tool.
func (s *Server) handleBBSRead(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	topicID, err := req.RequireFloat("topic_id")
	if err != nil {
		return mcp.NewToolResultError("topic_id is required and must be a number"), nil
	}

	limit := int(req.GetFloat("limit", 10))

	messages, err := s.db.GetMessages(int64(topicID), limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read messages: %v", err)), nil
	}

	if len(messages) == 0 {
		return mcp.NewToolResultText("No messages found"), nil
	}

	// Format messages as JSON
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal messages: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}
