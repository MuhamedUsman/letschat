package domain

import (
	"context"
	"time"
)

type Conversation struct {
	SenderID     string `json:"senderID"      db:"sender_id"`
	ReceiverID   string `json:"receiverID"    db:"receiver_id"`
	DisplayName  string `json:"displayName"  db:"display_name"`
	DisplayEmail string `json:"displayEmail"  db:"display_email"`
	// status of user other than the currently logged-in user, can be either sender or receiver
	LastOnline *time.Time `json:"lastOnline" db:"last_online"`
	// latest msg to display under user's name in TUI, only used on frontend side
	DisplayMsg string `json:"-"`
}

type ConversationService interface {
	CreateConversation(ctx context.Context, senderID, receiverID string) error
	GetConversations(ctx context.Context) ([]*Conversation, error)
	ConversationExists(ctx context.Context, senderID, receiverID string) (bool, error)
}

type ConversationRepository interface {
	CreateConversation(ctx context.Context, senderID, receiverID string) error
	GetConversations(ctx context.Context, usrID string) ([]*Conversation, error)
	ConversationExists(ctx context.Context, senderID, receiverID string) (bool, error)
}
