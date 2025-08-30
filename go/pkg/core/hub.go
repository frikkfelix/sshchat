package core

import (
	"context"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type Hub struct {
	sessions map[string]*Session
	channels map[string]*Channel

	register   chan *Session
	unregister chan string

	mu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	h := &Hub{
		sessions:   make(map[string]*Session),
		channels:   make(map[string]*Channel),
		register:   make(chan *Session, 16),
		unregister: make(chan string, 16),
		ctx:        ctx,
		cancel:     cancel,
	}

	h.createChannel("general", "General discussion")
	h.createChannel("random", "Random")

	return h
}

func (h *Hub) Run() {
	defer h.cleanup()
	for {
		select {
		case <-h.ctx.Done():
			return

		case session := <-h.register:
			h.handleRegister(session)
			go h.handleSession(session)

		case sessionID := <-h.unregister:
			h.handleUnregister(sessionID)
		}
	}
}

func (h *Hub) handleSession(session *Session) {
	defer func() {
		h.unregister <- session.ID
	}()

	for {
		select {
		case <-session.done:
			return

		case msg := <-session.inboundMessages():
			h.broadcastToChannel(msg)

		case cmd := <-session.commands:
			h.executeCommand(session, cmd)
		}
	}
}

func (h *Hub) handleRegister(session *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()

	select {
	case <-h.ctx.Done():
		session.Close()
		return
	default:
	}

	if _, exists := h.sessions[session.ID]; exists {
		log.Warn("Session %s already registered", session.ID)
		return
	}

	h.sessions[session.ID] = session

	go func() {
		_, cancel := context.WithTimeout(h.ctx, 1*time.Second)
		defer cancel()
		h.joinChannel(session, "general")
	}()
}

func (h *Hub) handleUnregister(sessionID string) {
	h.mu.Lock()
	session, exists := h.sessions[sessionID]
	if !exists {
		h.mu.Unlock()
		return
	}

	for _, channel := range h.channels {
		channel.RemoveSession(sessionID)
	}

	delete(h.sessions, sessionID)
	h.mu.Unlock()

	go func() {
		timer := time.NewTimer(100 * time.Millisecond)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-func() chan struct{} {
			done := make(chan struct{})
			go func() {
				close(session.outbox)
				close(done)
			}()
			return done
		}():
		}
	}()
}

func (h *Hub) broadcastToChannel(msg *Message) {
	h.mu.RLock()
	channel, exists := h.channels[msg.ChannelID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	channel.Broadcast(msg, h.sessions)
}

func (h *Hub) joinChannel(session *Session, channelName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	channel, exists := h.channels[channelName]
	if !exists {
		channel = h.createChannel(channelName, "")
	}

	if session.CurrentChannel != "" && session.CurrentChannel != channelName {
		if oldChannel, ok := h.channels[session.CurrentChannel]; ok {
			oldChannel.RemoveSession(session.ID)
		}
	}

	session.CurrentChannel = channelName
	channel.AddSession(session)

	for _, msg := range channel.GetRecentHistory(20) {
		h.sendToSession(session, msg)
	}
}

func (h *Hub) sendToSession(session *Session, msg *Message) {
	session.EnqueueOutbound(msg)
}

func (h *Hub) createChannel(name, topic string) *Channel {
	channel := NewChannel(name, topic)
	h.channels[name] = channel
	return channel
}

func (h *Hub) Shutdown() {
	h.cancel()
}

func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, session := range h.sessions {
		close(session.outbox)
	}
}

func (h *Hub) RegisterSession(session *Session) {
	h.register <- session
}
