package domain

import (
	"context"
	"regexp"
	"time"
)

// MsgOperation must not have math.MinInt8 value
type MsgOperation int

const (
	// CreateMsg indicates the sender has sent a new msg
	CreateMsg MsgOperation = iota
	// DeliveredMsg indicates the receiver has received the msg
	DeliveredMsg
	// DeliveredConfirmMsg indicates the sender's acknowledgment for the msg delivery
	DeliveredConfirmMsg
	// ReadMsg indicates the receiver has read the msg
	ReadMsg
	// ReadConfirmMsg indicates the sender's acknowledgment for the msg seen
	ReadConfirmMsg
	// DeleteMsg indicates the sender has deleted this msg
	DeleteMsg
	// DeleteConfirmMsg indicates the receiver's acknowledgment of the deleted message;
	// the receiving side will delete the msg, before sending this confirmation
	DeleteConfirmMsg
	// OnlineMsg indicates the user is online; a msg with this OP must not be persisted
	OnlineMsg
	// OfflineMsg indicates the user is offline; a msg with this OP must not be persisted
	OfflineMsg
	// TypingMsg indicates the user is typing; a msg with this OP must not be persisted
	TypingMsg
)

// Confirmation only be used on the frontend side
type Confirmation int

const (
	MsgDeliveredConfirmed Confirmation = iota + 1
	MsgReadConfirmed
)

var (
	rgxUUID = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$")
)

type Message struct {
	ID         string     `json:"id,omitempty"`
	SenderID   string     `json:"senderID,omitempty"   db:"sender_id"`
	ReceiverID string     `json:"receiverID,omitempty" db:"receiver_id"`
	Body       string     `json:"body,omitempty"`
	SentAt     *time.Time `json:"sent_at,omitempty"    db:"sent_at"`
	// used on the frontend side
	DeliveredAt  *time.Time   `db:"delivered_at"`
	ReadAt       *time.Time   `db:"read_at"`
	Confirmation Confirmation `json:"-"`
	Version      int          `json:"-"`
	Operation    MsgOperation `json:"operation"        db:"operation"`
}

type MsgChan chan *Message

type MessageService interface {
	PopulateMessage(m MessageSent, sndr *User) *Message
	ProcessSentMessages(ctx context.Context, m *Message) error
	GetUnDeliveredMessages(ctx context.Context, c MsgChan) error
	GetMessagesAsPage(ctx context.Context, c MsgChan, filter *Filter) (*Metadata, error)
	SaveMessage(ctx context.Context, m *Message) error
	//UpdateMessage(ctx context.Context, m *Message) error
	//DeleteMessage(ctx context.Context, mID string) error
}

type MessageRepository interface {
	GetByID(ctx context.Context, id string, op MsgOperation) (*Message, error)
	GetUnDeliveredMessages(ctx context.Context, rcvrID string, op MsgOperation, c MsgChan) error
	GetMessagesAsPage(ctx context.Context, rcvrID string, c MsgChan, filter *Filter) (*Metadata, error)
	InsertMessage(ctx context.Context, m *Message) error
	DeleteMessage(ctx context.Context, mID string) error
	DeleteMessageWithOperation(ctx context.Context, mID string, op MsgOperation) error
}

// DTO

type MessageSent struct {
	ID         *string      `json:"id"`
	ReceiverID string       `json:"receiverID"`
	Body       *string      `json:"body"`
	SentAt     *time.Time   `json:"sent_at"`
	Operation  MsgOperation `json:"operation"`
}

type LatestMsgBody struct {
	Body   *string    `db:"body"`
	SentAt *time.Time `db:"sent_at"`
}

// TODO: Update this logic
func (m MessageSent) ValidateMessageSent() *ErrValidation {
	return nil
}

func ValidateMessageRcvrID(id string, ev *ErrValidation) {
	ev.Evaluate(rgxUUID.MatchString(id), "receiverID", "Invalid receiver ID")
}

func ValidateMessageBody(body string, ev *ErrValidation) {
	ev.Evaluate(len(body) <= 5120, "body", "must be a max of 5120 bytes (5KB) long")
}
