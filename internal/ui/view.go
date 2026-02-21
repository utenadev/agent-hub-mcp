package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("211")).Bold(true)
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Background(lipgloss.Color("235"))
	dimStyle       = lipgloss.NewStyle().Faint(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	topicListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1)
	agentsListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1)
	messageStyle       = lipgloss.NewStyle().Padding(0, 1)
	senderStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	helpStyle          = lipgloss.NewStyle().Faint(true).Margin(1, 0)
	summaryStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238")).Padding(1)
	summaryHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("228")).Bold(true)
	summaryBadgeReal   = lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Background(lipgloss.Color("235")).Padding(0, 1)
	summaryBadgeMock   = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Background(lipgloss.Color("235")).Padding(0, 1)
	focusedBorderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("226"))
	onlineStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
	offlineStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	presenceHeader     = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
)

// Topic selector styles
var (
	topicSelectorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("226")).
			Padding(2).
			Width(40)
	topicSelectorTitle   = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Underline(true)
	topicSelectorItem   = lipgloss.NewStyle().Padding(0, 1)
	topicSelectorCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	topicSelectorHint   = lipgloss.NewStyle().Faint(true).MarginTop(1)
)

// View renders the model with 2-column vertical layout.
func (m Model) View() string {
	if m.Loading {
		return "Loading..."
	}

	if m.ErrorMessage != "" {
		return errorStyle.Render("Error: " + m.ErrorMessage)
	}

	// Topic selector modal
	if m.InputMode == ModeTopicSelect {
		return m.renderTopicSelector()
	}

	// Calculate dimensions
	leftWidth := m.Width / 3
	rightWidth := m.Width - leftWidth - 2 // Account for spacing
	contentHeight := m.Height - 4         // Reserve space for help text

	if leftWidth < 20 {
		leftWidth = 20
	}
	if rightWidth < 40 {
		rightWidth = 40
	}
	if contentHeight < 10 {
		contentHeight = 10
	}

	halfHeight := contentHeight / 2

	// Build left column: Topics (top) + Agents (bottom)
	topicsContent := m.renderTopicsPane()
	agentsContent := m.renderAgentsPane()

	var topicsPane, agentsPane string
	if m.FocusPane == PaneTopics {
		topicsPane = focusedBorderStyle.Width(leftWidth).Height(halfHeight).Render(topicsContent)
	} else {
		topicsPane = topicListStyle.Width(leftWidth).Height(halfHeight).Render(topicsContent)
	}

	if m.FocusPane == PaneAgents {
		agentsPane = focusedBorderStyle.Width(leftWidth).Height(halfHeight).Render(agentsContent)
	} else {
		agentsPane = agentsListStyle.Width(leftWidth).Height(halfHeight).Render(agentsContent)
	}

	leftColumn := lipgloss.JoinVertical(lipgloss.Left, topicsPane, agentsPane)

	// Build right column: Messages (top) + Summaries (bottom)
	messagesContent := m.renderMessagesPane()
	summariesContent := m.renderSummariesPane()

	var messagesPane, summariesPane string
	if m.FocusPane == PaneMessages {
		messagesPane = focusedBorderStyle.Width(rightWidth).Height(halfHeight).Render(messagesContent)
	} else {
		messagesPane = messageStyle.Width(rightWidth).Height(halfHeight).Render(messagesContent)
	}

	if m.FocusPane == PaneSummaries {
		summariesPane = focusedBorderStyle.Width(rightWidth).Height(halfHeight).Render(summariesContent)
	} else {
		summariesPane = summaryStyle.Width(rightWidth).Height(halfHeight).Render(summariesContent)
	}

	rightColumn := lipgloss.JoinVertical(lipgloss.Left, messagesPane, summariesPane)

	// Join columns horizontally
	layout := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	// Build help based on mode
	var bottom string
	if m.InputMode == ModePost {
		// Show sender and content inputs
		senderLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render("From: ")
		senderBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 1).
			Width(20).
			Render(m.SenderInput.View())
		contentLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render("Message: ")
		contentBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 1).
			Width(40).
			Render(m.TextInput.View())
		inputBox := lipgloss.JoinVertical(lipgloss.Left,
			senderLabel+senderBox,
			contentLabel+contentBox,
		)
		bottom = lipgloss.JoinVertical(lipgloss.Left, inputBox, helpStyle.Render("Enter: send | Esc: cancel"))
	} else {
		help := "h/j/k/l: nav | ←/→: focus | t: topics | [ / ]: summaries | r: refresh | p: post | q: quit"
		bottom = helpStyle.Render(help)
	}

	return lipgloss.JoinVertical(lipgloss.Left, layout, bottom)
}

// renderTopicsPane renders the topics list pane.
func (m Model) renderTopicsPane() string {
	var topicList strings.Builder
	topicList.WriteString(titleStyle.Render("Topics") + "\n\n")
	for i, topic := range m.Topics {
		if m.SelectedTopic != nil && topic.ID == m.SelectedTopic.ID {
			topicList.WriteString(selectedStyle.Render("▶ " + topic.Title))
		} else {
			topicList.WriteString("  " + topic.Title)
		}
		if i < len(m.Topics)-1 {
			topicList.WriteString("\n")
		}
	}

	return topicList.String()
}

// renderAgentsPane renders the agents/presence pane.
func (m Model) renderAgentsPane() string {
	var agentsList strings.Builder
	agentsList.WriteString(presenceHeader.Render("Agents") + "\n\n")

	if len(m.Presences) == 0 {
		agentsList.WriteString(dimStyle.Render("No agents online."))
		return agentsList.String()
	}

	for _, p := range m.Presences {
		var statusIndicator string
		if p.Status == "online" {
			statusIndicator = onlineStyle.Render("● ")
		} else {
			statusIndicator = offlineStyle.Render("○ ")
		}
		agentsList.WriteString(statusIndicator + p.Name + "\n")
		agentsList.WriteString(dimStyle.Render("  "+p.Role) + "\n")
	}

	return agentsList.String()
}

// renderMessagesPane renders the messages pane.
func (m Model) renderMessagesPane() string {
	var messageList strings.Builder
	if m.SelectedTopic != nil {
		messageList.WriteString(titleStyle.Render(m.SelectedTopic.Title) + "\n\n")
		for _, msg := range m.Messages {
			messageList.WriteString(senderStyle.Render(msg.Sender + ": "))
			messageList.WriteString(msg.Content + "\n")
		}
		if len(m.Messages) == 0 {
			messageList.WriteString(dimStyle.Render("No messages yet. Press 'p' to post."))
		}
	} else {
		messageList.WriteString(dimStyle.Render("Select a topic to view messages."))
	}
	return messageList.String()
}

// renderSummariesPane renders the summaries pane.
func (m Model) renderSummariesPane() string {
	var summaryList strings.Builder
	summaryList.WriteString(titleStyle.Render("Summaries") + "\n\n")

	if len(m.Summaries) == 0 {
		summaryList.WriteString(dimStyle.Render("No summaries available.\n\nOrchestrator will create\nsummaries periodically."))
		return summaryList.String()
	}

	// Show selected summary
	if m.SelectedSummaryIdx >= 0 && m.SelectedSummaryIdx < len(m.Summaries) {
		s := m.Summaries[m.SelectedSummaryIdx]

		// Header with badge
		summaryList.WriteString(summaryHeaderStyle.Render("📊 Summary #" + string(rune(len(m.Summaries)-m.SelectedSummaryIdx)) + "\n"))

		// Badge
		if s.IsMock {
			summaryList.WriteString(summaryBadgeMock.Render(" Mock ⚠️ "))
		} else {
			summaryList.WriteString(summaryBadgeReal.Render(" Gemini ✅ "))
		}
		summaryList.WriteString("\n\n")

		// Summary text (truncate if too long)
		summaryText := s.SummaryText
		if len(summaryText) > 500 {
			summaryText = summaryText[:497] + "..."
		}
		summaryList.WriteString(summaryText)

		// Navigation hint
		summaryList.WriteString("\n\n")
		summaryList.WriteString(dimStyle.Render("[" + string(rune('↓')) + "] older [" + string(rune('↑')) + "] newer"))
	}

	return summaryList.String()
}

// renderTopicSelector renders the topic selector modal.
func (m Model) renderTopicSelector() string {
	var sb strings.Builder

	// Title
	sb.WriteString(topicSelectorTitle.Render("Select Topic") + "\n\n")

	if len(m.Topics) == 0 {
		sb.WriteString(dimStyle.Render("No topics available.\nPress 'q' to quit."))
	} else {
		for i, topic := range m.Topics {
			if i == m.TopicSelectorIdx {
				sb.WriteString(topicSelectorCursor.Render("▶ " + topic.Title))
			} else {
				sb.WriteString("  " + topic.Title)
			}
			if i < len(m.Topics)-1 {
				sb.WriteString("\n")
			}
		}
	}

	// Hint
	sb.WriteString("\n")
	sb.WriteString(topicSelectorHint.Render("↑/k: up | ↓/j: down | Enter: select | Esc: cancel"))

	return topicSelectorStyle.Render(sb.String())
}
