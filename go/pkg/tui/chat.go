package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

const (
	textareaCharLimit     = 280
	minViewportHeight     = 3
	minTextareaInnerWidth = 10
	placeholderText       = "Type a message..."
)

type ChatModel struct {
	username string
	messages []Message
	viewport viewport.Model
	textarea textarea.Model
	session  ssh.Session
	width    int
	height   int
	ready    bool
}

type Message struct {
	Username string
	Text     string
	Time     string
}

type msgReceived Message

func NewChatModel(username string, session ssh.Session) *ChatModel {
	ta := textarea.New()
	configureTextarea(&ta)

	vp := viewport.New(3, 5)

	return &ChatModel{
		username: username,
		messages: []Message{},
		textarea: ta,
		viewport: vp,
		session:  session,
	}
}

func (m *ChatModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		listenForMessages(m.session),
	)
}

func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch message := msg.(type) {
	case tea.KeyMsg:
		switch message.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			trimmed := strings.TrimSpace(m.textarea.Value())
			if trimmed != "" {
				quitCmd := m.sendMessage(trimmed)
				m.textarea.Reset()
				cmds = append(cmds, quitCmd)
			}
		}

	case tea.WindowSizeMsg:
		m.handleWindowResize(message)
	case msgReceived:
		m.messages = append(m.messages, Message(message))
		m.updateViewport()
		cmds = append(cmds, listenForMessages(m.session))
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ChatModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf(
		"%s\n%s\n%s",
		m.HeaderView(),
		m.viewport.View(),
		inputBoxStyle.Render(m.textarea.View()),
	)
}

func (m *ChatModel) sendMessage(text string) tea.Cmd {
	if strings.HasPrefix(text, "/") {
		return m.handleCommand(text)
	}

	msg := Message{
		Username: m.username,
		Text:     text,
		Time:     timeNow(),
	}

	m.messages = append(m.messages, msg)
	m.updateViewport()

	// broadcast message to other users
	return nil
}

func (m *ChatModel) handleCommand(cmd string) tea.Cmd {
	var quitCmd tea.Cmd

	switch cmd {
	case "/help":
		m.messages = append(m.messages, Message{
			Username: "System",
			Text:     "Commands: /help, /clear, /quit",
			Time:     timeNow(),
		})
	case "/clear":
		m.messages = []Message{}
	case "/quit":
		quitCmd = tea.Quit
	default:
		m.messages = append(m.messages, Message{
			Username: "System",
			Text:     fmt.Sprintf("Unknown command: %s", cmd),
			Time:     timeNow(),
		})
	}
	m.updateViewport()
	return quitCmd
}

func (m *ChatModel) updateViewport() {
	var content strings.Builder
	for _, msg := range m.messages {
		content.WriteString(formatMessage(msg))
		content.WriteString("\n\n")
	}
	m.viewport.SetContent(content.String())
	m.viewport.GotoBottom()
}

func formatMessage(msg Message) string {
	user := userStyle.Render(msg.Username + ":")
	text := messageStyle.Render(msg.Text)
	return fmt.Sprintf("%s %s\n%s", user, timeStyle.Render(msg.Time), text)
}

func timeNow() string {
	return fmt.Sprintf("%02d:%02d",
		time.Now().Hour(),
		time.Now().Minute())
}

func listenForMessages(session ssh.Session) tea.Cmd {
	return func() tea.Msg {
		return nil
	}
}

func (m *ChatModel) handleWindowResize(message tea.WindowSizeMsg) {
	m.width = message.Width
	m.height = message.Height

	headerHeight := lipgloss.Height(m.HeaderView())
	inputHeight := lipgloss.Height(inputBoxStyle.Render(m.textarea.View()))

	vpHeight := message.Height - headerHeight - inputHeight
	if vpHeight < minViewportHeight {
		vpHeight = minViewportHeight
	}

	if !m.ready {
		m.viewport = viewport.New(message.Width, vpHeight)
		m.viewport.YPosition = headerHeight
		m.ready = true
	} else {
		m.viewport.Width = message.Width
		m.viewport.Height = vpHeight
		m.viewport.YPosition = headerHeight
	}

	m.viewport.Style = m.viewport.Style.PaddingLeft(1)

	inner := message.Width - inputBoxStyle.GetHorizontalPadding() - inputBoxStyle.GetHorizontalBorderSize()
	if inner < minTextareaInnerWidth {
		inner = minTextareaInnerWidth
	}
	m.textarea.SetWidth(inner)
	m.updateViewport()
}

func configureTextarea(ta *textarea.Model) {
	ta.Placeholder = placeholderText
	ta.Focus()
	ta.Prompt = ""
	ta.CharLimit = textareaCharLimit
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
