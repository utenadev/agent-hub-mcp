package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	titleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("211")).Bold(true)
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Background(lipgloss.Color("235"))
 dimStyle       = lipgloss.NewStyle().Faint(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	topicListStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1)
	messageStyle = lipgloss.NewStyle().Padding(0, 1)
	senderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	helpStyle     = lipgloss.NewStyle().Faint(true).Margin(1, 0)
	summaryStyle  = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1)
	summaryHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("228")).Bold(true)
	summaryBadgeReal  = lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Background(lipgloss.Color("235")).Padding(0, 1)
	summaryBadgeMock  = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Background(lipgloss.Color("235")).Padding(0, 1)
	focusedBorderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("226"))
)

// View renders the model.
func (m Model) View() string {
	if m.Loading {
		return "Loading..."
	}

	if m.ErrorMessage != "" {
		return errorStyle.Render("Error: " + m.ErrorMessage)
	}

	// Build the three panes
	topicsPane := m.renderTopicsPane()
	messagesPane := m.renderMessagesPane()
	summariesPane := m.renderSummariesPane()

	// Apply focused border to the active pane
	if m.FocusPane == PaneTopics {
		topicsPane = focusedBorderStyle.Width(25).Render(topicsPane)
	} else {
		topicsPane = topicListStyle.Width(25).Render(topicsPane)
	}

	if m.FocusPane == PaneMessages {
		messagesPane = focusedBorderStyle.Width(50).Render(messagesPane)
	} else {
		messagesPane = messageStyle.Width(50).Render(messagesPane)
	}

	if m.FocusPane == PaneSummaries {
		summariesPane = focusedBorderStyle.Width(35).Render(summariesPane)
	} else {
		summariesPane = summaryStyle.Width(35).Render(summariesPane)
	}

	// Join panes horizontally
	layout := lipgloss.JoinHorizontal(lipgloss.Left, topicsPane, messagesPane, summariesPane)

	// Build help
	help := "↑/k: up | ↓/j: down | Tab: focus | [ / ]: navigate summaries | r: refresh | p: post | q: quit"
	bottom := helpStyle.Render(help)

	return lipgloss.JoinVertical(lipgloss.Left, layout, bottom)
}

// renderTopicsPane renders the topics list pane.
func (m Model) renderTopicsPane() string {
	var topicList strings.Builder
	topicList.WriteString(titleStyle.Render("Topics\n\n"))
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

// renderMessagesPane renders the messages pane.
func (m Model) renderMessagesPane() string {
	var messageList strings.Builder
	if m.SelectedTopic != nil {
		messageList.WriteString(titleStyle.Render(m.SelectedTopic.Title + "\n\n"))
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
	summaryList.WriteString(titleStyle.Render("Summaries\n\n"))

	if len(m.Summaries) == 0 {
		summaryList.WriteString(dimStyle.Render("No summaries available.\n\nOrchestrator will create\nsummories periodically."))
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
		summaryList.WriteString(dimStyle.Render("["+string(rune('↓'))+"] older ["+string(rune('↑'))+"] newer"))
	}

	return summaryList.String()
}
