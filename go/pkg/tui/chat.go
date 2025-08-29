package tui

import (
	"fmt"
	"strings"

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

	input *InputController

	width  int
	height int
	ready  bool
}

type msgReceived core.Message

func NewModel(session *core.Session, h *core.Hub) *Model {
	return &Model{
		session:  session,
		hub:      h,
		messages: []core.Message{},
		input:    NewInputController(),
		viewport: viewport.New(80, 20),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.input.InitCmd(),
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

	switch v := msg.(type) {
	case tea.KeyMsg:
		if v.Type == tea.KeyCtrlC {
			m.session.Close()
			return m, tea.Quit
		}

		if cmd, handled := m.input.HandleKey(v, m.session); handled {
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		if m.input.Mode() == Normal {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(v)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		if cmd := m.input.Update(v); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.handleResize(v)
		inputHFrame, inputVFrame := inputBoxStyle.GetFrameSize()
		textAreaWidth := m.viewport.Width - inputHFrame
		if textAreaWidth < 0 {
			textAreaWidth = 0
		}
		m.input.OnResize(textAreaWidth, m.viewport.Height, lipgloss.Height(m.statusBar()), inputVFrame)

		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(v)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if cmd := m.input.Update(v); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case msgReceived:
		m.messages = append(m.messages, core.Message(v))
		m.updateViewport()
		cmds = append(cmds, m.listenForMessages())

	default:
		if cmd := m.input.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

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
		inputBoxStyle.Render(m.input.InlineView()),
		m.statusBar(),
	)

	return appFrameStyle.Render(content)
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
			Render(fmt.Sprintf("%s\n%s", timestamp, msg.Text))
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

	left := fmt.Sprintf("#%s | %s", channel, m.session.Username)
	right := ModeStyle(m.input.Mode()).Render(m.input.StatusLabel())

	totalWidth := m.viewport.Width
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)

	gap := totalWidth - leftW - rightW
	if gap < 1 {
		gap = 1
	}

	line := left + strings.Repeat(" ", gap) + right
	return statusStyle.Render(line)
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
	textAreaWidth := innerWidth - inputHFrame
	if textAreaWidth < 0 {
		textAreaWidth = 0
	}
	inputHeight := m.input.InlineHeight() + inputVFrame

	viewportHeight := innerHeight - headerHeight - statusHeight - inputHeight
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	m.viewport.Width = innerWidth
	m.viewport.Height = viewportHeight

	m.input.OnResize(textAreaWidth, viewportHeight, statusHeight, inputVFrame)
	m.updateViewport()

	if !m.ready {
		m.ready = true
	}
}
