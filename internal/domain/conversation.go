package domain

import (
	"context"
	"time"
)

type Conversation struct {
	SenderID      string `json:"senderID"      db:"sender_id"`
	SenderEmail   string `json:"senderEmail"   db:"sender_email"`
	SenderName    string `json:"senderName"    db:"sender_name"`
	ReceiverID    string `json:"receiverID"    db:"receiver_id"`
	ReceiverEmail string `json:"receiverEmail" db:"receiver_email"`
	ReceiverName  string `json:"receiverName"  db:"receiver_name"`
	// status of user other than the currently logged-in user, can be either sender or receiver
	LastOnline *time.Time `json:"lastOnline" db:"last_online"`
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
