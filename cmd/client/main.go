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

// ClientApp holds dependencies for testing
type ClientApp struct {
	HTTPClient *http.Client
	BaseURL    string
}

// NewClientApp creates a new ClientApp with defaults
func NewClientApp() *ClientApp {
	return &ClientApp{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		BaseURL:    "http://localhost:8080",
	}
}

func main() {
	app := NewClientApp()
	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// Run executes the client application
func (a *ClientApp) Run() error {
	// Get session ID from SSE
	resp, err := a.HTTPClient.Get(a.BaseURL + "/sse")
	if err != nil {
		return fmt.Errorf("failed to connect to SSE: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read SSE response: %w", err)
	}

	sessionID := string(body)
	fmt.Printf("SSE Response: %s\n", sessionID)

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
				"content":  "SSE経由で投稿しました！",
			},
		},
	}

	reqBody, _ := json.Marshal(postReq)
	fmt.Printf("Request: %s\n", string(reqBody))

	postResp, err := a.HTTPClient.Post(a.BaseURL+"/message", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to POST: %w", err)
	}
	defer postResp.Body.Close()

	respBody, _ := io.ReadAll(postResp.Body)
	fmt.Printf("Response: %s\n", string(respBody))

	return nil
}
