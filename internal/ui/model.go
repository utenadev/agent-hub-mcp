package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yklcs/agent-hub-mcp/internal/db"
)

// Model is the Bubble Tea model for the BBS dashboard.
type Model struct {
	db             *db.DB
	Topics         []db.Topic
	SelectedTopic  *db.Topic
	Messages       []db.Message
	Loading        bool
	ErrorMessage   string
	InputMode      InputMode
}

// InputMode represents the current input mode.
type InputMode int

const (
	ModeBrowse InputMode = iota
	ModePost
)

// TickMsg is sent periodically to refresh data.
type TickMsg time.Time

// TopicsLoadedMsg is sent when topics are loaded.
type TopicsLoadedMsg struct {
	Topics []db.Topic
	Error  error
}

// MessagesLoadedMsg is sent when messages are loaded.
type MessagesLoadedMsg struct {
	Messages []db.Message
	Error    error
}

// SelectTopicMsg is sent to select a topic.
type SelectTopicMsg int

// NewModel creates a new BBS dashboard model.
func NewModel(database *db.DB) Model {
	return Model{
		db:        database,
		Topics:    []db.Topic{},
		Messages:  []db.Message{},
		Loading:   true,
		InputMode: ModeBrowse,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return m.loadTopicsCmd()
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case TopicsLoadedMsg:
		m.Loading = false
		if msg.Error != nil {
			m.ErrorMessage = msg.Error.Error()
			return m, nil
		}
		m.Topics = msg.Topics
		// Auto-select first topic if available
		if len(m.Topics) > 0 && m.SelectedTopic == nil {
			return m, m.selectTopicCmd(m.Topics[0].ID)
		}
		return m, nil

	case MessagesLoadedMsg:
		m.Loading = false
		if msg.Error != nil {
			m.ErrorMessage = msg.Error.Error()
			return m, nil
		}
		m.Messages = msg.Messages
		return m, nil

	case TopicSelectedMsg:
		m.SelectedTopic = msg.Topic
		m.Messages = msg.Messages
		return m, nil

	case SelectTopicMsg:
		return m, m.selectTopicCmd(int(msg))

	case TickMsg:
		// Refresh data periodically
		return m, tea.Batch(m.loadTopicsCmd(), m.loadMessagesCmd())
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		// Navigate topics up
		if m.SelectedTopic != nil && len(m.Topics) > 0 {
			for i, t := range m.Topics {
				if t.ID == m.SelectedTopic.ID && i > 0 {
					return m, m.selectTopicCmd(m.Topics[i-1].ID)
				}
			}
		}
		return m, nil

	case "down", "j":
		// Navigate topics down
		if m.SelectedTopic != nil && len(m.Topics) > 0 {
			for i, t := range m.Topics {
				if t.ID == m.SelectedTopic.ID && i < len(m.Topics)-1 {
					return m, m.selectTopicCmd(m.Topics[i+1].ID)
				}
			}
		}
		return m, nil

	case "r":
		// Refresh
		return m, tea.Batch(m.loadTopicsCmd(), m.loadMessagesCmd())

	case "p":
		// Enter post mode
		m.InputMode = ModePost
		return m, nil
	}

	return m, nil
}

func (m Model) loadTopicsCmd() tea.Cmd {
	return func() tea.Msg {
		topics, err := m.db.ListTopics()
		return TopicsLoadedMsg{Topics: topics, Error: err}
	}
}

func (m Model) loadMessagesCmd() tea.Cmd {
	if m.SelectedTopic == nil {
		return nil
	}
	return func() tea.Msg {
		messages, err := m.db.GetMessages(int64(m.SelectedTopic.ID), 50)
		return MessagesLoadedMsg{Messages: messages, Error: err}
	}
}

func (m Model) selectTopicCmd(topicID int) tea.Cmd {
	return func() tea.Msg {
		// Find the topic
		var selected *db.Topic
		for _, t := range m.Topics {
			if t.ID == topicID {
				selected = &t
				break
			}
		}
		if selected == nil {
			return nil
		}
		// Load messages
		messages, err := m.db.GetMessages(int64(topicID), 50)
		if err != nil {
			return MessagesLoadedMsg{Messages: []db.Message{}, Error: err}
		}
		return TopicSelectedMsg{
			Topic:    selected,
			Messages: messages,
		}
	}
}

// TopicSelectedMsg is sent when a topic is selected with its messages.
type TopicSelectedMsg struct {
	Topic    *db.Topic
	Messages []db.Message
}
