package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frikkfelix/sshchat/go/pkg/core"
)

type Model struct {
	session *core.Session
	hub     *core.Hub

	messages []core.Message
	viewport viewport.Model
	textarea textarea.Model

	width  int
	height int
	ready  bool
}

type msgReceived core.Message

func NewModel(session *core.Session, h *core.Hub) *Model {
	ta := textarea.New()
	configureTextarea(&ta)

	return &Model{
		session:  session,
		hub:      h,
		messages: []core.Message{},
		textarea: ta,
		viewport: viewport.New(80, 20),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.listenForMessages(),
	)
}

func (m *Model) listenForMessages() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-m.session.Messages()
		if !ok {
			return tea.Quit
		}
		return msgReceived(*msg)
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.session.Close()
			return m, tea.Quit

		case tea.KeyEnter:
			text := strings.TrimSpace(m.textarea.Value())
			if text != "" {
				if strings.HasPrefix(text, "/") {
					m.handleCommand(text)
				} else {
					m.session.SendMessage(text)
				}
				m.textarea.Reset()
			}
		}

	case tea.WindowSizeMsg:
		m.handleResize(msg)

	case msgReceived:
		m.messages = append(m.messages, core.Message(msg))
		m.updateViewport()
		cmds = append(cmds, m.listenForMessages())
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	content := fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		m.headerView(),
		m.viewport.View(),
		m.statusBar(),
		inputBoxStyle.Render(m.textarea.View()))

	return appFrameStyle.Render(content)
}

func (m *Model) handleCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := core.Command{
		Name: parts[0][1:],
		Args: parts[1:],
	}

	m.session.SendCommand(cmd)
}

func (m *Model) updateViewport() {
	var content strings.Builder
	for _, msg := range m.messages {
		content.WriteString(m.formatMessage(msg))
		content.WriteString("\n\n")
	}

	wrapped := lipgloss.NewStyle().
		Width(m.viewport.Width).
		Render(content.String())

	m.viewport.SetContent(wrapped)
	m.viewport.GotoBottom()
}

func (m *Model) formatMessage(msg core.Message) string {
	timestamp := timeStyle.Render(fmt.Sprintf("%s", msg.Timestamp.Format("15:04")))

	key := msg.UserID

	if key == "" {
		key = msg.Username
	}

	user := UserStyle(key).Render(msg.Username)

	switch msg.Type {
	case core.MessageTypeSystem, core.MessageTypeJoin, core.MessageTypeLeave:
		return lipgloss.
			NewStyle().
			Render(fmt.Sprintf("%s %s", msg.Text, timestamp))
	default:
		return fmt.Sprintf("%s %s\n%s", user, timestamp, msg.Text)
	}
}

func (m *Model) headerView() string {
	headerContent := `
█▀ █▀ █ █ █▀▀ █ █ ▄▀█ ▀█▀
▄█ ▄█ █▀█ █▄▄ █▀█ █▀█  █
`
	return titleStyle.Render(headerContent)
}

func (m *Model) statusBar() string {
	channel := m.session.CurrentChannel
	if channel == "" {
		channel = "none"
	}
	return statusStyle.Render(fmt.Sprintf("#%s | %s", channel, m.session.Username))
}

func (m *Model) handleResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height

	outerHFrame, outerVFrame := appFrameStyle.GetFrameSize()
	innerWidth := msg.Width - outerHFrame
	if innerWidth < 0 {
		innerWidth = 0
	}
	innerHeight := msg.Height - outerVFrame
	if innerHeight < 0 {
		innerHeight = 0
	}

	headerHeight := lipgloss.Height(m.headerView())
	statusHeight := lipgloss.Height(statusStyle.Render(m.statusBar()))

	inputHFrame, inputVFrame := inputBoxStyle.GetFrameSize()
	inputHeight := m.textarea.Height() + inputVFrame

	viewportHeight := innerHeight - headerHeight - statusHeight - inputHeight
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	m.viewport.Width = innerWidth
	m.viewport.Height = viewportHeight

	textAreaWidth := innerWidth - inputHFrame
	if textAreaWidth < 0 {
		textAreaWidth = 0
	}
	m.textarea.SetWidth(textAreaWidth)
	m.updateViewport()

	if !m.ready {
		m.ready = true
	}
}

func configureTextarea(ta *textarea.Model) {
	ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.Prompt = ""
	ta.CharLimit = 280
	ta.Cursor.TextStyle.Bold(false)
	ta.Cursor.Style = cursorStyle
	ta.ShowLineNumbers = false
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.KeyMap.InsertNewline.SetEnabled(false)

	st, _ := textarea.DefaultStyles()
	st.Placeholder = st.Placeholder.
		Foreground(lipgloss.Color(textMuted)).
		Italic(true)
}
