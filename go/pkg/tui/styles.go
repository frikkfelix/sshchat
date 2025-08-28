package tui

import "github.com/charmbracelet/lipgloss"

const (
	colorAccent = "#7D56F4"
	textMuted   = "#666666"
	cursorColor = "#7D56F4"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorAccent)).
			PaddingLeft(2).
			PaddingBottom(1)

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(textMuted))

	userStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorAccent)).
			PaddingLeft(1)

	messageStyle = lipgloss.NewStyle().
			PaddingLeft(3)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorAccent)).
			Padding(0, 1)

	cursorStyle = lipgloss.
			NewStyle().
			Foreground(lipgloss.Color(cursorColor))
)
