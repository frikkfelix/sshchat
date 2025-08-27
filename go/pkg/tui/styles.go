package tui

import "github.com/charmbracelet/lipgloss"

const (
	titleColor  = "#7D56F4"
	headerColor = "#FFA500"
	textMuted   = "#666666"
	cursorColor = "#7D56F4"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(titleColor)).
			PaddingLeft(1).
			PaddingBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(headerColor)).
			PaddingLeft(1)

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(textMuted))

	userStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(titleColor)).
			PaddingLeft(1)

	messageStyle = lipgloss.NewStyle().
			PaddingLeft(3)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(titleColor)).
			Padding(0, 1)

	cursorStyle = lipgloss.
			NewStyle().
			Foreground(lipgloss.Color(cursorColor))
)
