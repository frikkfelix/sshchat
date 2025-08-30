package core

import (
	"fmt"
	"strings"
)

type Command struct {
	Name string
	Args []string
}

func (h *Hub) executeCommand(session *Session, cmd Command) {
	switch cmd.Name {
	case "help":
		h.cmdHelp(session)
	case "join", "j":
		if len(cmd.Args) > 0 {
			h.cmdJoin(session, cmd.Args[0])
		}
	case "list", "channels":
		h.cmdListChannels(session)
	case "users", "who":
		h.cmdListUsers(session)
	case "dm", "msg":
		if len(cmd.Args) >= 2 {
			h.cmdDirectMessage(session, cmd.Args[0], strings.Join(cmd.Args[1:], " "))
		}
	case "quit", "q", "q!":
		h.cmdQuit(session)
	default:
		h.sendToSession(session, NewMessage(
			MessageTypeError,
			"",
			"system",
			"System",
			fmt.Sprintf("Unknown command: /%s", cmd.Name),
		))
	}
}

func (h *Hub) cmdHelp(session *Session) {
	help := `Available commands:
/help - Show this help
/join <channel> - Join a channel
/list - List channels
/users - List users in current channel
/dm <user> <msg> - Send direct message
/quit - Exit`

	h.sendToSession(session, NewMessage(
		MessageTypeSystem,
		"",
		"system",
		"System",
		help,
	))
}

func (h *Hub) cmdJoin(session *Session, channel string) {
	h.joinChannel(session, channel)
}

func (h *Hub) cmdListChannels(session *Session) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var list []string
	for name, channel := range h.channels {
		list = append(list, fmt.Sprintf("#%s (%d users) - %s",
			name, channel.UserCount(), channel.Topic))
	}

	h.sendToSession(session, NewMessage(
		MessageTypeSystem,
		"",
		"system",
		"System",
		"Channels:\n"+strings.Join(list, "\n"),
	))
}

func (h *Hub) cmdListUsers(session *Session) {
	if session.CurrentChannel == "" {
		return
	}

	h.mu.RLock()
	channel := h.channels[session.CurrentChannel]
	h.mu.RUnlock()

	if channel == nil {
		return
	}

	channel.mu.RLock()
	var users []string
	for _, s := range channel.sessions {
		users = append(users, s.Username)
	}
	channel.mu.RUnlock()

	h.sendToSession(session, NewMessage(
		MessageTypeSystem,
		"",
		"system",
		"System",
		fmt.Sprintf("Users in #%s:\n%s", channel.Name, strings.Join(users, ", ")),
	))
}

func (h *Hub) cmdDirectMessage(session *Session, recipient, message string) {
	h.mu.RLock()
	var targetSession *Session
	for _, s := range h.sessions {
		if s.Username == recipient {
			targetSession = s
			break
		}
	}
	h.mu.RUnlock()

	if targetSession == nil {
		h.sendToSession(session, NewMessage(
			MessageTypeError,
			"",
			"system",
			"System",
			fmt.Sprintf("User %s not found", recipient),
		))
		return
	}

	dm := NewMessage(
		MessageTypePrivate,
		"",
		session.UserID,
		session.Username,
		message,
	)

	h.sendToSession(session, dm)
	h.sendToSession(targetSession, dm)
}

func (h *Hub) cmdQuit(session *Session) {
	session.Close()
}
