package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yklcs/agent-hub-mcp/internal/db"
)

func TestHandleBBSCreateTopic(t *testing.T) {
	// Use in-memory database for testing
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	server := NewServer(database)

	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid topic creation",
			args: map[string]interface{}{
				"title": "Test Topic",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			args: map[string]interface{}{
				"invalid": "value",
			},
			wantErr: true,
		},
		{
			name: "title is not string",
			args: map[string]interface{}{
				"title": 123,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}

			result, _ := server.handleBBSCreateTopic(context.Background(), req)

			if tt.wantErr {
				if !result.IsError {
					t.Error("expected error but got none")
				}
			} else {
				if result.IsError {
					tc, ok := mcp.AsTextContent(result.Content[0])
					if !ok {
						t.Errorf("unexpected error, cannot convert to TextContent")
					} else {
						t.Errorf("unexpected error result: %s", tc.Text)
					}
				}
			}
		})
	}
}

func TestHandleBBSPost(t *testing.T) {
	// Use in-memory database for testing
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Create a topic first
	topicID, err := database.CreateTopic("Test Topic")
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	server := NewServer(database)

	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid message post",
			args: map[string]interface{}{
				"topic_id": float64(topicID),
				"content":  "Hello, world!",
			},
			wantErr: false,
		},
		{
			name: "missing topic_id",
			args: map[string]interface{}{
				"content": "Hello",
			},
			wantErr: true,
		},
		{
			name: "missing content",
			args: map[string]interface{}{
				"topic_id": float64(topicID),
			},
			wantErr: true,
		},
		{
			name: "topic_id is not number",
			args: map[string]interface{}{
				"topic_id": "invalid",
				"content":  "Hello",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}

			result, _ := server.handleBBSPost(context.Background(), req)

			if tt.wantErr {
				if !result.IsError {
					t.Error("expected error but got none")
				}
			} else {
				if result.IsError {
					tc, ok := mcp.AsTextContent(result.Content[0])
					if !ok {
						t.Errorf("unexpected error, cannot convert to TextContent")
					} else {
						t.Errorf("unexpected error result: %s", tc.Text)
					}
				}
			}
		})
	}
}

func TestHandleBBSRead(t *testing.T) {
	// Use in-memory database for testing
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Create a topic and post some messages
	topicID, err := database.CreateTopic("Test Topic")
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	for i := 1; i <= 2; i++ {
		_, err := database.PostMessage(topicID, "alice", "Message")
		if err != nil {
			t.Fatalf("failed to post message: %v", err)
		}
	}

	server := NewServer(database)

	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid read",
			args: map[string]interface{}{
				"topic_id": float64(topicID),
			},
			wantErr: false,
		},
		{
			name: "valid read with limit",
			args: map[string]interface{}{
				"topic_id": float64(topicID),
				"limit":    float64(1),
			},
			wantErr: false,
		},
		{
			name: "missing topic_id",
			args: map[string]interface{}{
				"limit": float64(10),
			},
			wantErr: true,
		},
		{
			name: "topic_id is not number",
			args: map[string]interface{}{
				"topic_id": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}

			result, _ := server.handleBBSRead(context.Background(), req)

			if tt.wantErr {
				if !result.IsError {
					t.Error("expected error but got none")
				}
			} else {
				if result.IsError {
					tc, ok := mcp.AsTextContent(result.Content[0])
					if !ok {
						t.Errorf("unexpected error, cannot convert to TextContent")
					} else {
						t.Errorf("unexpected error result: %s", tc.Text)
					}
				}
			}
		})
	}
}
