package tui

import "github.com/charmbracelet/lipgloss"

const (
	textMuted = "#666666"

	colorRosewater = "#f5e0dc"
	colorFlamingo  = "#f2cdcd"
	colorPink      = "#f5c2e7"
	colorMauve     = "#cba6f7"
	colorRed       = "#f38ba8"
	colorMaroon    = "#eba0ac"
	colorPeach     = "#fab387"
	colorYellow    = "#f9e2af"
	colorGreen     = "#a6e3a1"
	colorTeal      = "#94e2d5"
	colorSky       = "#89dceb"
	colorSapphire  = "#74c7ec"
	colorBlue      = "#89b4fa"
	colorLavender  = "#b4befe"
	colorWhite     = "#FFFFFF"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorPeach)).
			PaddingBottom(1)

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(textMuted))

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorPeach)).
			Padding(0, 1)

	cursorStyle = lipgloss.
			NewStyle().
			Foreground(lipgloss.Color(colorPeach))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorWhite))

	appFrameStyle = lipgloss.NewStyle().PaddingBottom(1).PaddingLeft(1).PaddingRight(1)
)

var userPalette = []string{
	colorRed,
	colorYellow,
	colorGreen,
	colorTeal,
	colorBlue,
	colorMauve,
	colorPeach,
	colorSky,
	colorLavender,
	colorSapphire,
	colorPink,
	colorFlamingo,
	colorRosewater,
	colorMaroon,
	colorWhite,
}

func UserStyle(key string) lipgloss.Style {
	if key == "" {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorLavender))
	}
	idx := int(fnv32a(key) % uint32(len(userPalette)))
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(userPalette[idx]))
}

func ModeStyle(mode Mode) lipgloss.Style {
	switch mode {
	case Insert:
		return lipgloss.NewStyle().
			Background(lipgloss.Color(colorPeach)).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1)
	case Command:
		return lipgloss.NewStyle().
			Background(lipgloss.Color(colorBlue)).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1)
	default:
		return lipgloss.NewStyle().
			Background(lipgloss.Color(colorGreen)).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1)
	}
}

func fnv32a(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}
