package core

import (
	"sync"
)

type Channel struct {
	Name     string
	Topic    string
	sessions map[string]*Session
	history  []*Message
	mu       sync.RWMutex
}

func NewChannel(name, topic string) *Channel {
	return &Channel{
		Name:     name,
		Topic:    topic,
		sessions: make(map[string]*Session),
		history:  make([]*Message, 0, 100),
	}
}

func (c *Channel) AddSession(session *Session) {
	c.mu.Lock()
	c.sessions[session.ID] = session
	c.mu.Unlock()
}

func (c *Channel) RemoveSession(sessionID string) {
	c.mu.Lock()
	delete(c.sessions, sessionID)
	c.mu.Unlock()
}

func (c *Channel) Broadcast(msg *Message, allSessions map[string]*Session) {
	c.mu.Lock()

	c.history = append(c.history, msg)
	if len(c.history) > 100 {
		c.history = c.history[1:]
	}
	c.mu.Unlock()

	c.mu.RLock()
	sessionIDs := make([]string, 0, len(c.sessions))
	for id := range c.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	c.mu.RUnlock()

	for _, id := range sessionIDs {
		if session, ok := allSessions[id]; ok {
			session.EnqueueOutbound(msg)
		}
	}
}

func (c *Channel) GetRecentHistory(limit int) []*Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	start := 0
	if len(c.history) > limit {
		start = len(c.history) - limit
	}

	result := make([]*Message, len(c.history)-start)
	copy(result, c.history[start:])
	return result
}

func (c *Channel) UserCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.sessions)
}
