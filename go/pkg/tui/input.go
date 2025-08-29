package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frikkfelix/sshchat/go/pkg/core"
)

type Mode int

const (
	Insert Mode = iota
	Normal
)

type InputController struct {
	ta   textarea.Model
	mode Mode

	lastW        int
	lastViewport int
	lastStatusH  int
	lastInputVFr int
}

func NewInputController() *InputController {
	ta := textarea.New()
	configureTextarea(&ta)
	ic := &InputController{
		ta:   ta,
		mode: Insert,
	}
	ic.ta.Focus()
	return ic
}

func (i *InputController) InitCmd() tea.Cmd {
	return textarea.Blink
}

func (i *InputController) Update(msg tea.Msg) tea.Cmd {
	if i.mode == Insert {
		var cmd tea.Cmd
		i.ta, cmd = i.ta.Update(msg)
		return cmd
	}
	return nil
}

func (i *InputController) HandleKey(key tea.KeyMsg, session *core.Session) (tea.Cmd, bool) {
	if key.Type == tea.KeyCtrlC {
		return nil, false
	}

	switch i.mode {
	case Insert:
		switch key.Type {
		case tea.KeyEscape:
			i.enterNormal()
			return nil, true
		case tea.KeyEnter:
			text := strings.TrimSpace(i.ta.Value())

			if text == ":q!" {
				i.ta.Reset()
				session.Close()
				return tea.Quit, true
			}

			if text != "" {
				if strings.HasPrefix(text, "/") {
					parts := strings.Fields(text)
					if len(parts) > 0 {
						session.SendCommand(core.Command{Name: parts[0][1:], Args: parts[1:]})
					}
				} else {
					session.SendMessage(text)
				}
				i.ta.Reset()
			}
			return nil, true
		}
		return nil, false

	case Normal:
		switch key.String() {
		case "i", "enter":
			return i.enterInsert(), true
		}
		return nil, false

	}

	return nil, true
}

func (i *InputController) InlineView() string {
	return i.ta.View()
}

func (i *InputController) InlineHeight() int {
	return i.ta.Height()
}

func (i *InputController) StatusLabel() string {
	if i.mode == Insert {
		return "INSERT"
	}
	return "NORMAL"
}

func (i *InputController) Mode() Mode {
	return i.mode
}

func (i *InputController) EnterInsert() {
	i.enterInsert()
}

func (i *InputController) OnResize(textAreaWidth, innerViewportHeight, statusHeight, inputVFrame int) {
	i.lastW = textAreaWidth
	i.lastViewport = innerViewportHeight
	i.lastStatusH = statusHeight
	i.lastInputVFr = inputVFrame
	if textAreaWidth < 0 {
		textAreaWidth = 0
	}
	i.ta.SetWidth(textAreaWidth)
}

// internal
func (i *InputController) enterInsert() tea.Cmd {
	i.mode = Insert
	i.ta.Focus()
	return textarea.Blink
}

func (i *InputController) enterNormal() {
	i.mode = Normal
	i.ta.Blur()
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
