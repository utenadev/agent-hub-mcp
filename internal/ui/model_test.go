package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yklcs/agent-hub-mcp/internal/db"
)

// executeAllCmds executes all commands in a batch and updates the model.
func executeAllCmds(model Model, cmd tea.Cmd) Model {
	if cmd == nil {
		return model
	}
	msg := cmd()
	if msg == nil {
		return model
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			model = executeAllCmds(model, c)
		}
		return model
	}
	newModel, _ := model.Update(msg)
	return newModel.(Model)
}

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
		_, err := database.CreateTopic("Topic 1")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}
		_, err = database.CreateTopic("Topic 2")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}

		cmd := model.Init()
		model = executeAllCmds(model, cmd)

		if len(model.Topics) != 2 {
			t.Errorf("expected 2 topics, got %d", len(model.Topics))
		}
	})

	t.Run("SelectTopic selects a topic and loads messages", func(t *testing.T) {
		topicID, err := database.CreateTopic("Test Topic")
		if err != nil {
			t.Fatalf("failed to create topic: %v", err)
		}
		_, err = database.PostMessage(topicID, "alice", "Hello")
		if err != nil {
			t.Fatalf("failed to post message: %v", err)
		}

		cmd := model.Init()
		model = executeAllCmds(model, cmd)

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

		msg := SelectTopicMsg(foundTopic.ID)
		newModel, cmd := model.Update(msg)
		model = newModel.(Model)
		model = executeAllCmds(model, cmd)

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
