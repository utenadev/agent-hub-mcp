package db

import (
	"fmt"
	"os"
	"testing"
)

func TestDB(t *testing.T) {
	// Use in-memory database for testing
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	t.Run("CreateTopic", func(t *testing.T) {
		id, err := db.CreateTopic("Test Topic")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}
		if id == 0 {
			t.Error("expected non-zero topic id")
		}
	})

	t.Run("PostMessage", func(t *testing.T) {
		// Create a topic first
		topicID, err := db.CreateTopic("Test Topic for Messages")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}

		// Post a message
		msgID, err := db.PostMessage(topicID, "test_sender", "Hello, world!")
		if err != nil {
			t.Fatalf("failed to post message: %v", err)
		}
		if msgID == 0 {
			t.Error("expected non-zero message id")
		}
	})

	t.Run("GetMessages", func(t *testing.T) {
		// Create a topic and post some messages
		topicID, err := db.CreateTopic("Test Topic for GetMessages")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}

		// Post 3 messages
		for i := 1; i <= 3; i++ {
			_, err := db.PostMessage(topicID, "sender1", fmt.Sprintf("Message %d", i))
			if err != nil {
				t.Fatalf("failed to post message: %v", err)
			}
		}

		// Get messages with limit 2
		messages, err := db.GetMessages(topicID, 2)
		if err != nil {
			t.Fatalf("failed to get messages: %v", err)
		}
		if len(messages) != 2 {
			t.Errorf("expected 2 messages, got %d", len(messages))
		}

		// Get all messages (default limit)
		allMessages, err := db.GetMessages(topicID, 0)
		if err != nil {
			t.Fatalf("failed to get all messages: %v", err)
		}
		if len(allMessages) != 3 {
			t.Errorf("expected 3 messages, got %d", len(allMessages))
		}
	})

	t.Run("ListTopics", func(t *testing.T) {
		// Create some topics
		_, err := db.CreateTopic("Topic A")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}
		_, err = db.CreateTopic("Topic B")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}

		// List topics
		topics, err := db.ListTopics()
		if err != nil {
			t.Fatalf("failed to list topics: %v", err)
		}
		// Should have at least 2 topics (may have more from other tests)
		if len(topics) < 2 {
			t.Errorf("expected at least 2 topics, got %d", len(topics))
		}
	})
}

func TestDB_File(t *testing.T) {
	// Test with file-based database
	tmpFile := "/tmp/test-agent-hub.db"
	defer os.Remove(tmpFile)

	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create a topic
	id, err := db.CreateTopic("File Test Topic")
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero topic id")
	}

	// Close and reopen to verify persistence
	db.Close()

	db2, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db2.Close()

	// Verify topic exists
	topics, err := db2.ListTopics()
	if err != nil {
		t.Fatalf("failed to list topics: %v", err)
	}
	if len(topics) == 0 {
		t.Error("expected topics to persist after reopen")
	}
}
