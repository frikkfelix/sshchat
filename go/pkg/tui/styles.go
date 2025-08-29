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
			PaddingBottom(1)

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(textMuted))

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorAccent)).
			Padding(0, 1)

	cursorStyle = lipgloss.
			NewStyle().
			Foreground(lipgloss.Color(cursorColor))

	statusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(lipgloss.Color("#FFFFFF"))

	appFrameStyle = lipgloss.NewStyle().
			Padding(1, 2)
)

var userPalette = []string{
	"#E06C75",
	"#E5C07B",
	"#98C379",
	"#56B6C2",
	"#61AFEF",
	"#C678DD",
	"#D19A66",
	"#8BE9FD",
	"#50FA7B",
	"#BD93F9",
}

func UserStyle(key string) lipgloss.Style {
	if key == "" {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorAccent))
	}
	idx := int(fnv32a(key) % uint32(len(userPalette)))
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(userPalette[idx]))
}

func fnv32a(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}
