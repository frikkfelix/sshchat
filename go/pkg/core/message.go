package core

import (
	"time"

	"github.com/google/uuid"
)

type MessageType int

const (
	MessageTypeChat MessageType = iota
	MessageTypeSystem
	MessageTypeJoin
	MessageTypeLeave
	MessageTypePrivate
	MessageTypeError
)

type Message struct {
	ID        string      `json:"id"`
	Type      MessageType `json:"type"`
	ChannelID string      `json:"channel_id"`
	UserID    string      `json:"user_id"`
	Username  string      `json:"username"`
	Text      string      `json:"text"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewMessage(msgType MessageType, channelID, userID, username, text string) *Message {
	return &Message{
		ID:        uuid.NewString(),
		Type:      msgType,
		ChannelID: channelID,
		UserID:    userID,
		Username:  username,
		Text:      text,
		Timestamp: time.Now(),
	}
}
