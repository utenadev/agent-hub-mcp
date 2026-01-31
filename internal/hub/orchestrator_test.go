package hub

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/yklcs/agent-hub-mcp/internal/db"
)

func TestNewOrchestrator(t *testing.T) {
	// Create temporary database
	tmpDB, err := os.CreateTemp("", "test-orchestrator-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	database, err := db.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Test with default config
	orc := NewOrchestrator(database, nil)
	if orc == nil {
		t.Fatal("NewOrchestrator returned nil")
	}
	if orc.config.PollInterval != 5*time.Second {
		t.Errorf("Expected default PollInterval 5s, got %v", orc.config.PollInterval)
	}
	if orc.config.SummaryThreshold != 5 {
		t.Errorf("Expected default SummaryThreshold 5, got %d", orc.config.SummaryThreshold)
	}
}

func TestNewOrchestratorWithCustomConfig(t *testing.T) {
	tmpDB, err := os.CreateTemp("", "test-orchestrator-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	database, err := db.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	customConfig := &Config{
		PollInterval:     1 * time.Second,
		SummaryThreshold: 3,
		InactivityTimeout: 10 * time.Minute,
	}

	orc := NewOrchestrator(database, customConfig)
	if orc.config.PollInterval != 1*time.Second {
		t.Errorf("Expected PollInterval 1s, got %v", orc.config.PollInterval)
	}
	if orc.config.SummaryThreshold != 3 {
		t.Errorf("Expected SummaryThreshold 3, got %d", orc.config.SummaryThreshold)
	}
}

func TestInitializeTopics(t *testing.T) {
	tmpDB, err := os.CreateTemp("", "test-orchestrator-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	database, err := db.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Create some test topics
	topicID1, err := database.CreateTopic("Test Topic 1")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}
	topicID2, err := database.CreateTopic("Test Topic 2")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Add messages to first topic
	_, err = database.PostMessage(topicID1, "alice", "Hello")
	if err != nil {
		t.Fatalf("Failed to post message: %v", err)
	}
	_, err = database.PostMessage(topicID1, "bob", "Hi there")
	if err != nil {
		t.Fatalf("Failed to post message: %v", err)
	}

	// Create orchestrator and initialize
	orc := NewOrchestrator(database, nil)
	if err := orc.initializeTopics(); err != nil {
		t.Fatalf("initializeTopics failed: %v", err)
	}

	// Check that topics are tracked
	orc.mu.Lock()
	defer orc.mu.Unlock()

	if _, exists := orc.lastSeenMsgID[topicID1]; !exists {
		t.Error("Topic 1 not tracked in lastSeenMsgID")
	}
	if _, exists := orc.lastSeenMsgID[topicID2]; !exists {
		t.Error("Topic 2 not tracked in lastSeenMsgID")
	}

	// Topic 1 should have messages tracked
	if orc.lastSeenMsgID[topicID1] == 0 {
		t.Error("Topic 1 should have a tracked message ID")
	}

	// Topic 2 should have 0 as last seen (no messages)
	if orc.lastSeenMsgID[topicID2] != 0 {
		t.Errorf("Topic 2 should have 0 messages, got %d", orc.lastSeenMsgID[topicID2])
	}
}

func TestCheckTopicWithNewMessages(t *testing.T) {
	tmpDB, err := os.CreateTemp("", "test-orchestrator-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	database, err := db.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Create topic and initial message
	topicID, err := database.CreateTopic("Test Topic")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}
	_, err = database.PostMessage(topicID, "alice", "Initial message")
	if err != nil {
		t.Fatalf("Failed to post message: %v", err)
	}

	// Create orchestrator and initialize
	// Use high threshold to prevent summary generation during this test
	orc := NewOrchestrator(database, &Config{
		PollInterval:     100 * time.Millisecond,
		SummaryThreshold: 100, // High threshold to avoid triggering summary
		InactivityTimeout: 5 * time.Minute,
	})

	if err := orc.initializeTopics(); err != nil {
		t.Fatalf("initializeTopics failed: %v", err)
	}

	// Add new messages
	_, err = database.PostMessage(topicID, "bob", "New message 1")
	if err != nil {
		t.Fatalf("Failed to post message: %v", err)
	}
	_, err = database.PostMessage(topicID, "charlie", "New message 2")
	if err != nil {
		t.Fatalf("Failed to post message: %v", err)
	}

	// Check for new messages
	if err := orc.checkTopic(topicID); err != nil {
		t.Fatalf("checkTopic failed: %v", err)
	}

	orc.mu.Lock()
	count := orc.topicMsgCount[topicID]
	orc.mu.Unlock()

	if count != 2 {
		t.Errorf("Expected 2 new messages tracked, got %d", count)
	}
}

func TestMockSummarizer(t *testing.T) {
	tmpDB, err := os.CreateTemp("", "test-orchestrator-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	database, err := db.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	orc := NewOrchestrator(database, nil)

	messages := []db.Message{
		{ID: 1, Sender: "alice", Content: "Hello"},
		{ID: 2, Sender: "bob", Content: "Hi"},
		{ID: 3, Sender: "alice", Content: "How are you?"},
	}

	summary := orc.mockSummarizer(messages)

	if summary == "" {
		t.Error("Summary should not be empty")
	}

	// Check that summary contains expected elements
	// Note: Not checking exact format since it's a mock
	if len(summary) < 10 {
		t.Errorf("Summary too short: %q", summary)
	}
}

func TestGenerateSummary(t *testing.T) {
	tmpDB, err := os.CreateTemp("", "test-orchestrator-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	database, err := db.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Create topic and add messages
	topicID, err := database.CreateTopic("Test Topic")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err = database.PostMessage(topicID, "alice", fmt.Sprintf("Message %d", i))
		if err != nil {
			t.Fatalf("Failed to post message: %v", err)
		}
	}

	// Create orchestrator and set up state
	orc := NewOrchestrator(database, nil)
	orc.mu.Lock()
	orc.topicMsgCount[topicID] = 5 // Set to threshold to trigger summary
	orc.mu.Unlock()

	// Generate summary
	if err := orc.generateSummary(topicID); err != nil {
		t.Fatalf("generateSummary failed: %v", err)
	}

	// Verify summary was posted
	messages, err := database.GetMessages(topicID, 10)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	// Should have original 3 messages + 1 summary
	if len(messages) < 4 {
		t.Errorf("Expected at least 4 messages, got %d", len(messages))
	}

	// Find the summary message
	foundSummary := false
	for _, msg := range messages {
		if msg.Sender == "orchestrator" {
			foundSummary = true
			break
		}
	}

	if !foundSummary {
		t.Error("Summary message not found in topic")
	}

	// Verify counter was reset
	orc.mu.Lock()
	count := orc.topicMsgCount[topicID]
	orc.mu.Unlock()

	if count != 0 {
		t.Errorf("Expected message count to be reset to 0, got %d", count)
	}
}
