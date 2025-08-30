package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frikkfelix/sshchat/go/pkg/core"
)

type Mode int

const (
	Insert Mode = iota
	Normal
	Command // dedicated ":" command-line mode using textinput
)

type InputController struct {
	ta         textarea.Model
	mode       Mode
	cmdActive  bool
	cmd        textinput.Model
	cmdCurrent string
	lastW      int
}

func NewInputController() *InputController {
	ta := textarea.New()
	configureTextarea(&ta)

	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = ""
	ti.Blur()

	ic := &InputController{
		ta:   ta,
		mode: Insert,
		cmd:  ti,
	}
	ic.ta.Focus()
	return ic
}

func (i *InputController) InitCmd() tea.Cmd {
	return textarea.Blink
}

func (i *InputController) Update(msg tea.Msg) tea.Cmd {
	if i.cmdActive {
		var cmd tea.Cmd
		i.cmd, cmd = i.cmd.Update(msg)
		i.syncCommandPrompt()
		return cmd
	}
	if i.mode == Insert {
		var cmd tea.Cmd
		i.ta, cmd = i.ta.Update(msg)
		return cmd
	}
	return nil
}

func (i *InputController) HandleKey(k tea.KeyMsg, session *core.Session) (tea.Cmd, bool) {
	if k.Type == tea.KeyCtrlC {
		return nil, false
	}
	if i.cmdActive {
		switch k.Type {
		case tea.KeyEscape:
			i.exitCommandBar()
			return nil, true
		case tea.KeyEnter:
			return i.executeCommand(session), true
		default:
			// normal typing goes through Update so it shows up
			return nil, false
		}
	}

	switch i.mode {
	case Normal:
		if k.String() == ":" {
			i.openCommandBar()
			return textinput.Blink, true
		}
		if k.String() == "i" || k.String() == "enter" {
			return i.enterInsert(), true
		}
		return nil, false

	case Insert:
		if k.String() == ":" && strings.TrimSpace(i.ta.Value()) == "" {
			i.openCommandBar()
			return textinput.Blink, true
		}
		switch k.Type {
		case tea.KeyEscape:
			i.enterNormal()
			return nil, true
		case tea.KeyEnter:
			text := strings.TrimSpace(i.ta.Value())
			if text == ":q!" || text == ":q" {
				i.ta.Reset()
				session.Close()
				return tea.Quit, true
			}
			if strings.HasPrefix(text, "/") {
				parts := strings.Fields(text)
				if len(parts) > 0 {
					session.SendCommand(core.Command{
						Name: strings.TrimPrefix(parts[0], "/"),
						Args: parts[1:],
					})
				}
				i.ta.Reset()
				return nil, true
			}
			if text != "" {
				session.SendMessage(text)
				i.ta.Reset()
			}
			return nil, true
		}
		return nil, false
	}
	return nil, false
}

func (i *InputController) InlineView() string {
	return i.ta.View()
}

func (i *InputController) CommandInline(avail int) string {
	if !i.cmdActive || avail <= 0 {
		return ""
	}
	i.syncCommandPrompt()

	promptW := lipgloss.Width(i.cmd.Prompt)
	fieldW := avail - promptW
	if fieldW < 1 {
		fieldW = 1
	}
	i.cmd.Width = fieldW

	return i.cmd.View()
}

func (i *InputController) InlineHeight() int {
	return i.ta.Height()
}

func (i *InputController) StatusLabel() string {
	if i.cmdActive {
		return "COMMAND"
	}
	if i.mode == Insert {
		return "INSERT"
	}
	return "NORMAL"
}

func (i *InputController) Mode() Mode {
	if i.cmdActive {
		return Command
	}
	return i.mode
}

func (i *InputController) EnterInsert() tea.Cmd {
	return i.enterInsert()
}

func (i *InputController) OnResize(textAreaWidth, _innerViewportHeight, _statusHeight, _inputVFrame int) {
	i.lastW = textAreaWidth
	if textAreaWidth < 0 {
		textAreaWidth = 0
	}
	i.ta.SetWidth(textAreaWidth)
}

func (i *InputController) enterInsert() tea.Cmd {
	i.mode = Insert
	i.ta.Focus()
	return textarea.Blink
}

func (i *InputController) enterNormal() {
	i.mode = Normal
	i.ta.Blur()
}

func (i *InputController) openCommandBar() {
	i.cmdActive = true
	i.cmd.Focus()
	i.cmd.SetValue("")
	i.cmd.Prompt = ":" // colored when a word appears
	i.cmdCurrent = ""
}

func (i *InputController) exitCommandBar() {
	i.cmdActive = false
	i.cmd.Blur()
	i.cmd.SetValue("")
	i.cmd.Prompt = ""
	i.cmdCurrent = ""
}

func (i *InputController) executeCommand(session *core.Session) tea.Cmd {
	raw := strings.TrimSpace(i.cmd.Value())
	if raw == "" {
		i.exitCommandBar()
		return nil
	}
	parts := strings.Fields(raw)
	name := parts[0]
	args := parts[1:]
	if name == "q" || name == "q!" || raw == ":q" || raw == ":q!" {
		i.exitCommandBar()
		session.Close()
		return tea.Quit
	}
	session.SendCommand(core.Command{Name: name, Args: args})
	i.exitCommandBar()
	return nil
}

func (i *InputController) syncCommandPrompt() {
	raw := strings.TrimLeft(i.cmd.Value(), " \t")
	word, _ := splitFirstWord(raw)
	i.cmdCurrent = word
	i.cmd.Prompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color(textMuted)).
		Render(":")
}

func splitFirstWord(s string) (word string, rest string) {
	s = strings.TrimLeft(s, " \t")
	if s == "" {
		return "", ""
	}
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return "", ""
	}
	word = parts[0]
	rest = strings.TrimPrefix(s, parts[0])
	rest = strings.TrimLeft(rest, " \t")
	return word, rest
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
