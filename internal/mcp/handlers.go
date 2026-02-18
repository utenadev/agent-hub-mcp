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

	sender := s.getSender()

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
	sender := s.getSender()

	unreadCount, err := s.db.CountUnreadMessages(sender)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to count unread messages: %v", err)), nil
	}

	if err := s.db.UpdateAgentCheckTime(sender); err != nil {
		fmt.Printf("Warning: failed to update check time: %v\n", err)
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
		result += "\n\n【重要：連携ガイドライン】BBSに未読メッセージがあります。リソース `guidelines://agent-collaboration` に基づき、現在の作業を保存し、最優先で `bbs_read` を実行してください。確認後は `update_status` で状況を報告してください。"
	}

	return mcp.NewToolResultText(result), nil
}

// handleUpdateStatus handles the update_status tool.
func (s *Server) handleUpdateStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sender := s.getSender()

	status, err := req.RequireString("status")
	if err != nil {
		return mcp.NewToolResultError("status is required and must be a string"), nil
	}

	topicIDFloat := req.GetFloat("topic_id", 0)
	var topicID *int64
	if topicIDFloat > 0 {
		tid := int64(topicIDFloat)
		topicID = &tid
	}

	if err := s.db.UpdateAgentStatus(sender, status, topicID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update status: %v", err)), nil
	}

	topicInfo := ""
	if topicID != nil {
		topicInfo = fmt.Sprintf(" on topic %d", *topicID)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Status updated: %s%s", status, topicInfo)), nil
}

// handleRegisterAgent handles the bbs_register_agent tool.
func (s *Server) handleRegisterAgent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required and must be a string"), nil
	}

	role, err := req.RequireString("role")
	if err != nil {
		return mcp.NewToolResultError("role is required and must be a string"), nil
	}

	status := req.GetString("status", "online")

	topicIDFloat := req.GetFloat("topic_id", 0)
	var topicID *int64
	if topicIDFloat > 0 {
		tid := int64(topicIDFloat)
		topicID = &tid
	}

	if err := s.db.UpsertAgentPresence(name, role); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to register agent: %v", err)), nil
	}

	s.CurrentSender = name
	s.DefaultRole = role

	if err := s.db.UpdateAgentStatus(name, status, topicID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to set initial status: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Agent registered: name=%s, role=%s, status=%s", name, role, status)), nil
}
