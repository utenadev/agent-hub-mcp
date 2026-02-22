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
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	model := NewModel(database)

	t.Run("Initial state", func(t *testing.T) {
		if model.Topics == nil {
			t.Error("expected Topics to be initialized")
		}
		if model.Messages == nil {
			t.Error("expected Messages to be initialized")
		}
	})
}

func TestInitLoadsTopics(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	database.CreateTopic("Topic 1")
	database.CreateTopic("Topic 2")

	model := NewModel(database)
	cmd := model.Init()
	model = executeAllCmds(model, cmd)

	if len(model.Topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(model.Topics))
	}
}

func TestTopicSelectorNavigation(t *testing.T) {
	database, _ := db.Open(":memory:")
	defer database.Close()
	database.CreateTopic("Topic 1")
	database.CreateTopic("Topic 2")
	database.CreateTopic("Topic 3")

	// Test up arrow from index 2
	model := NewModel(database)
	model.InputMode = ModeTopicSelect
	model.Topics, _ = database.ListTopics()
	model.TopicSelectorIdx = 2

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	m := newModel.(Model)
	if m.TopicSelectorIdx != 1 {
		t.Errorf("up: expected 1, got %d", m.TopicSelectorIdx)
	}

	// Test down arrow from index 0
	model2 := NewModel(database)
	model2.InputMode = ModeTopicSelect
	model2.Topics, _ = database.ListTopics()
	model2.TopicSelectorIdx = 0

	newModel2, _ := model2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := newModel2.(Model)
	if m2.TopicSelectorIdx != 1 {
		t.Errorf("down: expected 1, got %d", m2.TopicSelectorIdx)
	}
}

func TestTopicSelectorEsc(t *testing.T) {
	database, _ := db.Open(":memory:")
	defer database.Close()

	model := NewModel(database)
	model.InputMode = ModeTopicSelect

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := newModel.(Model)
	if m.InputMode != ModeBrowse {
		t.Errorf("expected ModeBrowse, got %d", m.InputMode)
	}
}

func TestPostModeEnterAndEsc(t *testing.T) {
	database, _ := db.Open(":memory:")
	defer database.Close()

	topicID, _ := database.CreateTopic("Test Topic")

	// Test p key enters post mode
	model := NewModel(database)
	model.SelectedTopic = &db.Topic{ID: int(topicID), Title: "Test Topic"}
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m := newModel.(Model)
	if m.InputMode != ModePost {
		t.Errorf("p key: expected ModePost, got %d", m.InputMode)
	}

	// Test escape exits post mode
	model2 := NewModel(database)
	model2.InputMode = ModePost
	newModel2, _ := model2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := newModel2.(Model)
	if m2.InputMode != ModeBrowse {
		t.Errorf("esc: expected ModeBrowse, got %d", m2.InputMode)
	}
}

func TestFocusPaneNavigation(t *testing.T) {
	database, _ := db.Open(":memory:")
	defer database.Close()

	// Test Tab
	model := NewModel(database)
	model.FocusPane = PaneTopics
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	m := newModel.(Model)
	if m.FocusPane != PaneAgents {
		t.Errorf("Tab: expected PaneAgents, got %d", m.FocusPane)
	}

	// Test Shift+Tab
	model2 := NewModel(database)
	model2.FocusPane = PaneAgents
	newModel2, _ := model2.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m2 := newModel2.(Model)
	if m2.FocusPane != PaneTopics {
		t.Errorf("Shift+Tab: expected PaneTopics, got %d", m2.FocusPane)
	}

	// Test right arrow
	model3 := NewModel(database)
	model3.FocusPane = PaneTopics
	newModel3, _ := model3.Update(tea.KeyMsg{Type: tea.KeyRight})
	m3 := newModel3.(Model)
	if m3.FocusPane != PaneAgents {
		t.Errorf("right: expected PaneAgents, got %d", m3.FocusPane)
	}

	// Test left arrow
	model4 := NewModel(database)
	model4.FocusPane = PaneAgents
	newModel4, _ := model4.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m4 := newModel4.(Model)
	if m4.FocusPane != PaneTopics {
		t.Errorf("left: expected PaneTopics, got %d", m4.FocusPane)
	}

	// Test h key
	model5 := NewModel(database)
	model5.FocusPane = PaneAgents
	newModel5, _ := model5.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m5 := newModel5.(Model)
	if m5.FocusPane != PaneTopics {
		t.Errorf("h: expected PaneTopics, got %d", m5.FocusPane)
	}

	// Test l key
	model6 := NewModel(database)
	model6.FocusPane = PaneTopics
	newModel6, _ := model6.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m6 := newModel6.(Model)
	if m6.FocusPane != PaneAgents {
		t.Errorf("l: expected PaneAgents, got %d", m6.FocusPane)
	}
}

func TestSummariesNavigation(t *testing.T) {
	database, _ := db.Open(":memory:")
	defer database.Close()

	model := NewModel(database)
	model.FocusPane = PaneSummaries
	model.Summaries = []db.TopicSummary{
		{ID: 1, TopicID: 1, SummaryText: "Summary 1", IsMock: false},
		{ID: 2, TopicID: 1, SummaryText: "Summary 2", IsMock: true},
	}
	model.SelectedSummaryIdx = 1

	// Test ]
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	m := newModel.(Model)
	if m.SelectedSummaryIdx != 0 {
		t.Errorf("]: expected 0, got %d", m.SelectedSummaryIdx)
	}

	// Test [
	model2 := NewModel(database)
	model2.FocusPane = PaneSummaries
	model2.Summaries = []db.TopicSummary{
		{ID: 1, TopicID: 1, SummaryText: "Summary 1", IsMock: false},
		{ID: 2, TopicID: 1, SummaryText: "Summary 2", IsMock: true},
	}
	model2.SelectedSummaryIdx = 0
	newModel2, _ := model2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")})
	m2 := newModel2.(Model)
	if m2.SelectedSummaryIdx != 1 {
		t.Errorf("[: expected 1, got %d", m2.SelectedSummaryIdx)
	}
}

func TestTopicNavigation(t *testing.T) {
	database, _ := db.Open(":memory:")
	defer database.Close()

	topicID1, _ := database.CreateTopic("Topic 1")
	topicID2, _ := database.CreateTopic("Topic 2")

	// Test j navigates to next topic
	model := NewModel(database)
	model.Topics = []db.Topic{
		{ID: int(topicID1), Title: "Topic 1"},
		{ID: int(topicID2), Title: "Topic 2"},
	}
	model.SelectedTopic = &model.Topics[0]
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m := newModel.(Model)
	if cmd != nil {
		m = executeAllCmds(m, cmd)
	}
	if m.SelectedTopic == nil || m.SelectedTopic.ID != int(topicID2) {
		t.Errorf("j: expected topic ID %d, got %v", topicID2, m.SelectedTopic)
	}

	// Test t opens topic selector
	model2 := NewModel(database)
	model2.InputMode = ModeBrowse
	newModel2, _ := model2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m2 := newModel2.(Model)
	if m2.InputMode != ModeTopicSelect {
		t.Errorf("t: expected ModeTopicSelect, got %d", m2.InputMode)
	}
	if m2.TopicSelectorIdx != 0 {
		t.Errorf("t: expected TopicSelectorIdx 0, got %d", m2.TopicSelectorIdx)
	}
}

func TestRefresh(t *testing.T) {
	database, _ := db.Open(":memory:")
	defer database.Close()

	model := NewModel(database)
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	_ = newModel.(Model)
	if cmd == nil {
		t.Error("expected refresh commands")
	}
}
