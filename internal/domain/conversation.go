package domain

import (
	"context"
	"time"
)

type Conversation struct {
	SenderID   string `json:"-"              db:"sender_id"`
	ReceiverID string `json:"-"              db:"receiver_id"`
	// Below given attributes will only be used on TUI (frontend side)
	UserID    string `json:"userID"          db:"user_id"`
	Username  string `json:"username"        db:"username"`
	UserEmail string `json:"userEmail"       db:"user_email"`
	// status of user other than the currently logged-in user, can be either sender or receiver
	LastOnline *time.Time `json:"lastOnline" db:"last_online"`
	// latest msg to display under user's name in TUI, only used on frontend side
	LatestMsg       *string    `json:"-"`
	LatestMsgSentAt *time.Time `json:"-"`
}

type ConversationService interface {
	CreateConversation(ctx context.Context, senderID, receiverID string) (bool, error)
	GetConversations(ctx context.Context) ([]*Conversation, error)
	ConversationExists(ctx context.Context, senderID, receiverID string) (bool, error)
}

type ConversationRepository interface {
	CreateConversation(ctx context.Context, senderID, receiverID string) (bool, error)
	GetConversations(ctx context.Context, usrID string) ([]*Conversation, error)
	ConversationExists(ctx context.Context, senderID, receiverID string) (bool, error)
}
