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
)

// View renders the model.
func (m Model) View() string {
	if m.Loading {
		return "Loading..."
	}

	if m.ErrorMessage != "" {
		return errorStyle.Render("Error: " + m.ErrorMessage)
	}

	// Build topic list
	topicList := strings.Builder{}
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

	// Build message list
	messageList := strings.Builder{}
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

	// Build help
	help := "↑/k: up | ↓/j: down | r: refresh | p: post | q: quit"

	// Layout
	left := topicListStyle.Render(topicList.String())
	right := messageStyle.Render(messageList.String())
	layout := lipgloss.JoinHorizontal(lipgloss.Left, left, right)
	bottom := helpStyle.Render(help)

	return lipgloss.JoinVertical(lipgloss.Left, layout, bottom)
}
