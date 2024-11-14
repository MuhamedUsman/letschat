package repository

import (
	"database/sql"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"log"
)

type LocalMessageRepository struct {
	db *DB
}

func NewLocalMessageRepository(db *DB) LocalMessageRepository {
	return LocalMessageRepository{db}
}

type LatestMsgs map[string]*domain.LatestMsgBody

func (r LocalMessageRepository) GetLatestMsgBodyForConvos(cui ...string) (LatestMsgs, error) {
	// cui convoUsrId
	query := `
		SELECT body, sent_at
		FROM message
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY sent_at DESC
	`
	msgs := make(LatestMsgs, len(cui))
	for _, id := range cui {
		var msg domain.LatestMsgBody
		var t string
		if err := r.db.QueryRow(query, id).Scan(&msg.Body, &t); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue // Ignore and continue, as there may be no messages for some chats
			}
			return nil, err
		}
		msg.SentAt, _ = parseTime(&t)
		msgs[id] = &msg
	}
	return msgs, nil
}

func (r LocalMessageRepository) GetMsgByID(id string) (*domain.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, body, sent_at, delivered_at, read_at, confirmation, version
		FROM message
		WHERE id = $1
	`
	var msg domain.Message
	var SentAt, DeliveredAt, ReadAt *string
	args := []any{&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Body, &SentAt, &DeliveredAt, &ReadAt, &msg.Confirmation, &msg.Version}
	if err := r.db.QueryRow(query, id).Scan(args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	msg.SentAt, _ = parseTime(SentAt)
	msg.DeliveredAt, _ = parseTime(DeliveredAt)
	msg.ReadAt, _ = parseTime(ReadAt)
	return &msg, nil
}

func (r LocalMessageRepository) SaveMsg(msg *domain.Message) error {
	query := `
		INSERT INTO message (id, sender_id, receiver_id, body, sent_at, delivered_at, read_at, confirmation)
		VALUES (:id, :sender_id, :receiver_id, :body, :sent_at, :delivered_at, :read_at, :confirmation)
	`
	_, err := r.db.NamedExec(query, msg)
	return err
}

func (r LocalMessageRepository) UpdateMsg(msg *domain.Message) error {
	query := `
		UPDATE message 
		SET delivered_at = :delivered_at, read_at = :read_at, confirmation = :confirmation, version = version + 1
		WHERE id = :id AND version = :version
	`
	res, err := r.db.NamedExec(query, msg)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrEditConflict
	}
	return nil
}

func (r LocalMessageRepository) DeleteMsg(id string) error {
	query := `
		DELETE FROM message WHERE id = $1
	`
	_, err := r.db.Exec(query, id)
	return err
}

func (r LocalMessageRepository) DeleteAllForSenderAndReceiver(senderId, receiverId string) error {
	query := `
		DELETE FROM message 
        WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
	`
	res, err := r.db.Exec(query, senderId, receiverId)
	log.Println(res)
	return err
}

func (r LocalMessageRepository) GetMsgsAsPage(
	sen string,
	fil domain.Filter,
) ([]*domain.Message, *domain.Metadata, error) {
	query := `
		SELECT COUNT(*) OVER(), id, sender_id, receiver_id, body, sent_at, delivered_at, read_at, confirmation, version
		FROM message
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY sent_at DESC
		LIMIT $2
	    OFFSET $3
		`
	args := []any{sen, fil.Limit(), fil.Offset()}
	rows, _ := r.db.Query(query, args...)
	var TotalRows int
	msgs := make([]*domain.Message, 0)
	for rows.Next() {
		var m domain.Message
		var SentAt, DeliveredAt, ReadAt *string
		args = []any{&TotalRows, &m.ID, &m.SenderID, &m.ReceiverID, &m.Body, &SentAt, &DeliveredAt, &ReadAt, &m.Confirmation, &m.Version}
		if err := rows.Scan(args...); err != nil {
			return nil, &domain.Metadata{}, err
		}
		m.SentAt, _ = parseTime(SentAt)
		m.DeliveredAt, _ = parseTime(DeliveredAt)
		m.ReadAt, _ = parseTime(ReadAt)
		msgs = append(msgs, &m)
	}
	metadata := domain.CalculateMetadata(TotalRows, fil.PageSize, fil.Page)
	return msgs, &metadata, nil
}
