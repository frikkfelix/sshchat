package tui

func (m *ChatModel) HeaderView() string {
	headerContent := `
█▀ █▀ █ █ █▀▀ █ █ ▄▀█ ▀█▀
▄█ ▄█ █▀█ █▄▄ █▀█ █▀█  █
`
	return titleStyle.Render(headerContent)
}
