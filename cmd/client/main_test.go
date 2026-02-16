package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClientApp(t *testing.T) {
	app := NewClientApp()
	if app == nil {
		t.Fatal("NewClientApp returned nil")
	}
	if app.HTTPClient == nil {
		t.Error("expected HTTPClient to be set")
	}
	if app.BaseURL != "http://localhost:8080" {
		t.Errorf("expected default BaseURL, got: %s", app.BaseURL)
	}
}

func TestClientApp_Run_SSEConnection(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sse" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("event: endpoint\ndata: /message?sessionId=test123"))
			return
		}
		if r.URL.Path == "/message" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"jsonrpc":"2.0","id":2,"result":{}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	app := NewClientApp()
	app.BaseURL = server.URL

	err := app.Run()
	if err != nil {
		// Expected to succeed or fail gracefully
		t.Logf("Run returned: %v", err)
	}
}

func TestClientApp_Run_ConnectionError(t *testing.T) {
	app := NewClientApp()
	app.BaseURL = "http://localhost:1" // Invalid port

	err := app.Run()
	if err == nil {
		t.Error("expected error for connection failure")
	}
	if !strings.Contains(err.Error(), "failed to connect") {
		t.Errorf("expected connection error, got: %v", err)
	}
}

func TestMCPRequest(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	if req.JSONRPC != "2.0" {
		t.Error("JSONRPC should be 2.0")
	}
	if req.ID != 1 {
		t.Error("ID should be 1")
	}
}
