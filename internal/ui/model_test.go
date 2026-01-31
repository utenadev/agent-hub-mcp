package ui

import (
	"testing"

	"github.com/yklcs/agent-hub-mcp/internal/db"
)

func TestModel(t *testing.T) {
	// Use in-memory database for testing
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Create initial model
	model := NewModel(database)

	t.Run("Initial state", func(t *testing.T) {
		if model.Topics == nil {
			t.Error("expected Topics to be initialized")
		}
		if model.Messages == nil {
			t.Error("expected Messages to be initialized")
		}
		// SelectedTopic can be nil initially
	})

	t.Run("Init command loads topics", func(t *testing.T) {
		// Create some topics
		_, err := database.CreateTopic("Topic 1")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}
		_, err = database.CreateTopic("Topic 2")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}

		// Init should load topics
		cmd := model.Init()
		if cmd != nil {
			// Wait for the command to complete
			msg := cmd()
			if msg != nil {
				newModel, _ := model.Update(msg)
				model = newModel.(Model)
			}
		}

		if len(model.Topics) != 2 {
			t.Errorf("expected 2 topics, got %d", len(model.Topics))
		}
	})

	t.Run("SelectTopic selects a topic and loads messages", func(t *testing.T) {
		// Create a topic with messages
		topicID, err := database.CreateTopic("Test Topic")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}
		_, err = database.PostMessage(topicID, "alice", "Hello")
		if err != nil {
			t.Fatalf("failed to post message: %v", err)
		}

		// Reload topics to get the new one
		cmd := model.Init()
		if cmd != nil {
			msg := cmd()
			if msg != nil {
				newModel, _ := model.Update(msg)
				model = newModel.(Model)
			}
		}

		// Find the topic in the list
		var foundTopic *db.Topic
		for i, t := range model.Topics {
			if t.ID == int(topicID) {
				foundTopic = &model.Topics[i]
				break
			}
		}
		if foundTopic == nil {
			t.Fatal("topic not found in list")
		}

		// Select the topic using the topic ID
		msg := SelectTopicMsg(foundTopic.ID)
		newModel, cmd := model.Update(msg)
		model = newModel.(Model)

		// Execute the command to load messages
		if cmd != nil {
			resultMsg := cmd()
			if resultMsg != nil {
				newModel, _ = model.Update(resultMsg)
				model = newModel.(Model)
			}
		}

		if model.SelectedTopic == nil {
			t.Error("expected SelectedTopic to be set")
		} else if model.SelectedTopic.ID != foundTopic.ID {
			t.Errorf("expected topic ID %d, got %d", foundTopic.ID, model.SelectedTopic.ID)
		}
		if len(model.Messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(model.Messages))
		}
	})
}
