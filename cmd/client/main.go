package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func main() {
	// Get session ID from SSE
	resp, err := http.Get("http://localhost:8080/sse")
	if err != nil {
		fmt.Printf("Failed to connect to SSE: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read SSE response: %v\n", err)
		return
	}

	// Parse sessionId from: event: endpoint\ndata: /message?sessionId=xxx
	sessionID := string(body)
	fmt.Printf("SSE Response: %s\n", sessionID)

	// For now, extract sessionId manually or use a known format
	// Let's try a direct POST with a fresh session

	// First, establish SSE connection and get a real session
	// For simplicity, let's use tools/list which might work without init
	time.Sleep(100 * time.Millisecond)

	// Try to post a message
	postReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: ToolCallParams{
			Name: "bbs_post",
			Arguments: map[string]interface{}{
				"topic_id": 8,
				"content":  "SSE経由で投稿しました！🚀",
			},
		},
	}

	reqBody, _ := json.Marshal(postReq)
	fmt.Printf("Request: %s\n", string(reqBody))

	// Try posting to the message endpoint directly
	postResp, err := http.Post("http://localhost:8080/message", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Failed to POST: %v\n", err)
		return
	}
	defer postResp.Body.Close()

	respBody, _ := io.ReadAll(postResp.Body)
	fmt.Printf("Response: %s\n", string(respBody))
}
