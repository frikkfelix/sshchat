package core

import (
	"github.com/charmbracelet/ssh"
)

type Session struct {
	ID             string
	UserID         string
	Username       string
	CurrentChannel string
	inbox          chan *Message
	outbox         chan *Message
	commands       chan Command
	done           chan struct{}
}

func NewSession(sshSession ssh.Session) *Session {
	fingerprint := sshSession.Context().Value("fingerprint").(string)

	username := sshSession.User()
	if username == "" {
		username = "anonymous-" + fingerprint[:8]
	}

	return &Session{
		ID:       fingerprint,
		UserID:   fingerprint,
		Username: username,
		inbox:    make(chan *Message, 64),
		outbox:   make(chan *Message, 64),
		commands: make(chan Command, 16),
		done:     make(chan struct{}),
	}
}

func (s *Session) SendMessage(text string) {
	if s.CurrentChannel == "" {
		return
	}

	msg := NewMessage(
		MessageTypeChat,
		s.CurrentChannel,
		s.UserID,
		s.Username,
		text,
	)

	select {
	case s.inbox <- msg:
	default:
		// buffer full
	}
}

func (s *Session) SendCommand(cmd Command) {
	select {
	case s.commands <- cmd:
	default:
		// buffer full
	}
}

func (s *Session) EnqueueOutbound(msg *Message) {
	select {
	case s.outbox <- msg:
	default:
	}
}

func (s *Session) Messages() <-chan *Message {
	return s.outbox
}

func (s *Session) inboundMessages() <-chan *Message {
	return s.inbox
}

func (s *Session) Close() {
	select {
	case <-s.done:
		return
	default:
		close(s.done)
	}
}
