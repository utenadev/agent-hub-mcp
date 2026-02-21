package ui

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yklcs/agent-hub-mcp/internal/db"
)

// Model is the Bubble Tea model for the Agent Hub dashboard.
type Model struct {
	db                 *db.DB
	Topics             []db.Topic
	SelectedTopic      *db.Topic
	Messages           []db.Message
	Summaries          []db.TopicSummary
	Presences          []db.AgentPresence
	SelectedSummaryIdx int
	FocusPane          FocusPane
	Loading            bool
	ErrorMessage       string
	InputMode          InputMode
	TextInput          textinput.Model
	SenderInput        textinput.Model
	Width              int
	Height             int
	TopicSelectorIdx    int
	PostField          int // 0: sender, 1: content
}

// InputMode represents the current input mode.
type InputMode int

const (
	ModeBrowse InputMode = iota
	ModePost
	ModeTopicSelect
)

// FocusPane represents which pane has focus.
type FocusPane int

const (
	PaneTopics FocusPane = iota
	PaneAgents
	PaneMessages
	PaneSummaries
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

// SummariesLoadedMsg is sent when summaries are loaded.
type SummariesLoadedMsg struct {
	Summaries []db.TopicSummary
	Error     error
}

// PresenceLoadedMsg is sent when agent presence is loaded.
type PresenceLoadedMsg struct {
	Presences []db.AgentPresence
	Error     error
}

// SelectTopicMsg is sent to select a topic.
type SelectTopicMsg int

// NewModel creates a new Agent Hub dashboard model.
func NewModel(database *db.DB) Model {
	ti := textinput.New()
	ti.Placeholder = "Type your message..."
	ti.CharLimit = 1000
	ti.Width = 48

	senderInput := textinput.New()
	senderInput.Placeholder = "Your name"
	senderInput.CharLimit = 100
	senderInput.Width = 20
	// Set default sender
	defaultSender := os.Getenv("BBS_AGENT_ID")
	if defaultSender == "" {
		defaultSender = "Human"
	}
	senderInput.SetValue(defaultSender)

	return Model{
		db:                 database,
		Topics:             []db.Topic{},
		Messages:           []db.Message{},
		Summaries:          []db.TopicSummary{},
		SelectedSummaryIdx: 0,
		FocusPane:          PaneTopics,
		Loading:            true,
		InputMode:          ModeTopicSelect,
		TextInput:          ti,
		SenderInput:        senderInput,
		Width:              80,
		Height:             24,
		TopicSelectorIdx:   0,
		PostField:          1, // Start with content field
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadTopicsCmd(),
		m.tickCmd(),
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

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

	case SummariesLoadedMsg:
		if msg.Error != nil {
			return m, nil
		}
		m.Summaries = msg.Summaries
		m.SelectedSummaryIdx = 0
		return m, nil

	case PresenceLoadedMsg:
		if msg.Error != nil {
			return m, nil
		}
		m.Presences = msg.Presences
		return m, nil

	case TopicSelectedMsg:
		m.SelectedTopic = msg.Topic
		m.Messages = msg.Messages
		// Load summaries for the selected topic
		return m, m.loadSummariesCmd()

	case SelectTopicMsg:
		return m, m.selectTopicCmd(int(msg))

	case TickMsg:
		return m, tea.Batch(
			m.loadTopicsCmd(),
			m.loadMessagesCmd(),
			m.loadPresenceCmd(),
			m.tickCmd(),
		)

	case MessagePostedMsg:
		if msg.Error != nil {
			m.ErrorMessage = msg.Error.Error()
			return m, nil
		}
		// Auto-refresh messages after successful post
		return m, m.loadMessagesCmd()
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle input mode first - text input takes priority
	if m.InputMode == ModePost {
		switch msg.String() {
		case "esc":
			// Exit post mode
			m.InputMode = ModeBrowse
			m.TextInput.Reset()
			m.TextInput.Blur()
			return m, nil
		case "enter":
			// Submit message
			if m.TextInput.Value() != "" && m.SelectedTopic != nil {
				content := m.TextInput.Value()
				sender := m.SenderInput.Value()
				if sender == "" {
					sender = "Human"
				}
				m.TextInput.Reset()
				m.InputMode = ModeBrowse
				m.TextInput.Blur()
				return m, m.postMessageCmd(sender, content)
			}
			return m, nil
		default:
			// Pass other keys to text input
			var cmd tea.Cmd
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
	}

	// Topic selector mode
	if m.InputMode == ModeTopicSelect {
		switch msg.String() {
		case "esc":
			m.InputMode = ModeBrowse
			return m, nil
		case "enter":
			if len(m.Topics) > m.TopicSelectorIdx {
				m.InputMode = ModeBrowse
				return m, m.selectTopicCmd(m.Topics[m.TopicSelectorIdx].ID)
			}
			return m, nil
		case "up", "k":
			if m.TopicSelectorIdx > 0 {
				m.TopicSelectorIdx--
			}
			return m, nil
		case "down", "j":
			if m.TopicSelectorIdx < len(m.Topics)-1 {
				m.TopicSelectorIdx++
			}
			return m, nil
		}
		return m, nil
	}

	// Browse mode key handling
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

	case "t":
		// Open topic selector
		m.InputMode = ModeTopicSelect
		m.TopicSelectorIdx = 0
		return m, nil

	case "p":
		// Enter post mode
		if m.SelectedTopic != nil {
			m.InputMode = ModePost
			m.TextInput.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case "tab":
		// Cycle focus forward between panes
		switch m.FocusPane {
		case PaneTopics:
			m.FocusPane = PaneAgents
		case PaneAgents:
			m.FocusPane = PaneMessages
		case PaneMessages:
			m.FocusPane = PaneSummaries
		case PaneSummaries:
			m.FocusPane = PaneTopics
		}
		return m, nil

	case "shift+tab":
		// Cycle focus backward between panes
		switch m.FocusPane {
		case PaneTopics:
			m.FocusPane = PaneSummaries
		case PaneAgents:
			m.FocusPane = PaneTopics
		case PaneMessages:
			m.FocusPane = PaneAgents
		case PaneSummaries:
			m.FocusPane = PaneMessages
		}
		return m, nil

	case "right", "l":
		// Cycle focus forward (same as Tab)
		switch m.FocusPane {
		case PaneTopics:
			m.FocusPane = PaneAgents
		case PaneAgents:
			m.FocusPane = PaneMessages
		case PaneMessages:
			m.FocusPane = PaneSummaries
		case PaneSummaries:
			m.FocusPane = PaneTopics
		}
		return m, nil

	case "left", "h":
		// Cycle focus backward (same as Shift+Tab)
		switch m.FocusPane {
		case PaneTopics:
			m.FocusPane = PaneSummaries
		case PaneAgents:
			m.FocusPane = PaneTopics
		case PaneMessages:
			m.FocusPane = PaneAgents
		case PaneSummaries:
			m.FocusPane = PaneMessages
		}
		return m, nil

	case "]":
		// Navigate to newer summary (towards index 0)
		if m.FocusPane == PaneSummaries && len(m.Summaries) > 0 {
			if m.SelectedSummaryIdx > 0 {
				m.SelectedSummaryIdx--
			}
		}
		return m, nil

	case "[":
		// Navigate to older summary (towards end of list)
		if m.FocusPane == PaneSummaries && len(m.Summaries) > 0 {
			if m.SelectedSummaryIdx < len(m.Summaries)-1 {
				m.SelectedSummaryIdx++
			}
		}
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

func (m Model) loadSummariesCmd() tea.Cmd {
	if m.SelectedTopic == nil {
		return nil
	}
	return func() tea.Msg {
		// Get all summaries for this topic (most recent first)
		summaries, err := m.db.GetSummariesByTopic(int64(m.SelectedTopic.ID))
		if err != nil {
			return SummariesLoadedMsg{Summaries: []db.TopicSummary{}, Error: err}
		}
		return SummariesLoadedMsg{Summaries: summaries}
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

// MessagePostedMsg is sent when a message is posted.
type MessagePostedMsg struct {
	Error error
}

func (m Model) postMessageCmd(sender, content string) tea.Cmd {
	return func() tea.Msg {
		if m.SelectedTopic == nil {
			return MessagePostedMsg{Error: nil}
		}
		_, err := m.db.PostMessage(int64(m.SelectedTopic.ID), sender, content)
		return MessagePostedMsg{Error: err}
	}
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m Model) loadPresenceCmd() tea.Cmd {
	return func() tea.Msg {
		presences, err := m.db.ListAllAgentPresence()
		return PresenceLoadedMsg{Presences: presences, Error: err}
	}
}
