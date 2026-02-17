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

	// Use the server's default sender (configured via -sender flag or BBS_AGENT_ID env var)
	sender := s.DefaultSender
	if sender == "" {
		sender = "unknown"
	}

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

// handleCheckHubStatus handles the check_hub_status tool.
func (s *Server) handleCheckHubStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Update last check time
	if err := s.db.UpdateAgentCheckTime(s.DefaultSender); err != nil {
		// Non-fatal: continue with status check
		fmt.Printf("Warning: failed to update check time: %v\n", err)
	}

	// Get unread message count
	unreadCount, err := s.db.CountUnreadMessages(s.DefaultSender)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to count unread messages: %v", err)), nil
	}

	// Get all agents' presence
	presences, err := s.db.ListAllAgentPresence()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get team presence: %v", err)), nil
	}

	// Build response
	response := map[string]interface{}{
		"has_new_activity": unreadCount > 0,
		"unread_count":     unreadCount,
		"team_presence":    presences,
	}

	// Format as JSON
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	result := string(data)

	// Inject prompt if there are unread messages
	if unreadCount > 0 {
		result += "\n\n【システム通知】BBSに未読メッセージがあります。作業の区切りで `bbs_read` を実行して指示を確認してください。"
	}

	return mcp.NewToolResultText(result), nil
}

// handleUpdateStatus handles the update_status tool.
func (s *Server) handleUpdateStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status, err := req.RequireString("status")
	if err != nil {
		return mcp.NewToolResultError("status is required and must be a string"), nil
	}

	// Get optional topic_id
	topicIDFloat := req.GetFloat("topic_id", 0)
	var topicID *int64
	if topicIDFloat > 0 {
		tid := int64(topicIDFloat)
		topicID = &tid
	}

	// Update status in database
	if err := s.db.UpdateAgentStatus(s.DefaultSender, status, topicID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update status: %v", err)), nil
	}

	var topicInfo string
	if topicID != nil {
		topicInfo = fmt.Sprintf(" on topic %d", *topicID)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Status updated: %s%s", status, topicInfo)), nil
}
