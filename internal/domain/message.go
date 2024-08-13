package domain

import (
	"context"
	"regexp"
	"time"
)

type MessageOperation int

const (
	CreateMsg MessageOperation = iota
	UpdateMsg
	DeleteMsg
	GetMoreMsg // Only sent through sub endpoint to get more messages
)

var (
	rgxUUID = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$")
)

type Message struct {
	ID          string           `json:"id"`
	SenderID    string           `json:"senderID,omitempty"    db:"sender_id"`
	ReceiverID  string           `json:"receiverID,omitempty"  db:"receiver_id"`
	Body        string           `json:"body,omitempty"`
	SentAt      *time.Time       `json:"sentAt,omitempty"      db:"sent_at"`
	DeliveredAt *time.Time       `json:"deliveredAt,omitempty" db:"delivered_at"`
	ReadAt      *time.Time       `json:"readAt,omitempty"      db:"read_at"`
	Version     int              `json:"-"`
	Operation   MessageOperation `json:"operation"`
}

type MsgChan chan *Message

type MessageService interface {
	PopulateMessage(ctx context.Context, m MessageSent) *Message
	ProcessSentMessages(ctx context.Context, m *Message) error
	GetUnreadMessages(ctx context.Context, c MsgChan) error
	GetMessagesAsPage(ctx context.Context, c MsgChan, filter *Filter) (*Metadata, error)
	SaveMessage(ctx context.Context, m *Message) error
	UpdateMessage(ctx context.Context, m *Message) error
	DeleteMessage(ctx context.Context, mID string) error
}

type MessageRepository interface {
	GetByID(ctx context.Context, id string) (*Message, error)
	GetUnreadMessages(ctx context.Context, rcvrID string, c MsgChan) error
	GetMessagesAsPage(ctx context.Context, rcvrID string, c MsgChan, filter *Filter) (*Metadata, error)
	InsertMessage(ctx context.Context, m *Message) error
	UpdateMessage(ctx context.Context, m *Message) error
	DeleteMessage(ctx context.Context, mID string) error
}

// DTO

type MessageSent struct {
	ID          *string          `json:"id"`
	ReceiverID  string           `json:"receiverID"`
	Body        *string          `json:"body"`
	SentAt      *time.Time       `json:"sentAt"`
	DeliveredAt *time.Time       `json:"deliveredAt"`
	ReadAt      *time.Time       `json:"readAt"`
	Operation   MessageOperation `json:"operation"`
}

func (m MessageSent) ValidateMessageSent() *ErrValidation {
	ev := NewErrValidation()
	ValidateMessageRcvrID(m.ReceiverID, ev)
	if m.Operation == CreateMsg {
		ValidateMessageBody(*m.Body, ev)
		ev.Evaluate(m.SentAt != nil, "sentAt", "required for 0(CreateMsg) operation")
	}
	if m.Operation == UpdateMsg {
		ev.Evaluate(m.Body == nil, "body", "nil required for 1(UpdateMsg) operation)")
		ev.Evaluate(m.ID != nil, "id", "required for 1(UpdateMsg) operation")
		if m.DeliveredAt == nil && m.ReadAt == nil {
			ev.AddError("deliveredAt", "(Optional) required to 1(UpdateMsg)")
			ev.AddError("readAt", "(Optional) required to 1(UpdateMsg)")
		}
	}
	if m.Operation == DeleteMsg && m.ID == nil {
		ev.AddError("id", "required for 2(DeleteMsg) operation")
	}
	if ev.HasErrors() {
		return ev
	}
	return nil
}

func ValidateMessageRcvrID(id string, ev *ErrValidation) {
	ev.Evaluate(rgxUUID.MatchString(id), "receiverID", "Invalid receiver ID")
}

func ValidateMessageBody(body string, ev *ErrValidation) {
	ev.Evaluate(len(body) <= 5120, "body", "must be a max of 5120 bytes (5KB) long")
}
