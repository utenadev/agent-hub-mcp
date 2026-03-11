package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yklcs/agent-hub-mcp/internal/db"
)

func TestServeStreamableHTTP(t *testing.T) {
	// Setup
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	srv := NewServer(database, "test-sender", "test-role")

	mux := http.NewServeMux()
	handler := srv.NewStreamableHTTPHandler()
	mux.Handle("/mcp/", handler)

	// Start a real test server to avoid hanging with NewRecorder
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Use a client with a very short timeout since we only care about headers
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	req, err := http.NewRequest("GET", ts.URL+"/mcp/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We expect the server to send headers and then keep the connection open for SSE.
	// We want to verify the headers.
	resp, err := client.Do(req)
	
	// If it timed out but we got headers, that's okay for this test.
	// But usually headers are sent immediately.
	if err != nil {
		t.Fatalf("failed to get response: %v", err)
	}
	defer resp.Body.Close()

	// Assert Status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}

	// Assert Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("expected content-type text/event-stream, got %s", contentType)
	}

	// Assert CORS Headers (This is expected to fail initially)
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got '%s'", allowOrigin)
	}
}
