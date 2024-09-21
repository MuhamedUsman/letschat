package repository

import (
	"database/sql"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"time"
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
		SELECT * FROM message WHERE id = $1
	`
	var msg domain.Message
	if err := r.db.QueryRowx(query, id).StructScan(&msg); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return &msg, nil
}

func (r LocalMessageRepository) SaveMsg(msg *domain.Message) error {
	query := `
		INSERT INTO message (id, sender_id, receiver_id, body, sent_at, delivered_at, read_at)
		VALUES (:id, :sender_id, :receiver_id, :body, :sent_at, :delivered_at, :read_at)
	`
	_, err := r.db.NamedExec(query, msg)
	return err
}

func (r LocalMessageRepository) UpdateMsg(msg *domain.Message) error {
	query := `
		UPDATE message 
		SET delivered_at = :delivered_at, read_at = :read_at, version = version + 1
		WHERE id = :id AND version = :version
	`
	res, err := r.db.NamedExec(query, msg)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return domain.ErrEditConflict
	}
	return nil
}

func (r LocalMessageRepository) DeleteMsg(id string) error {
	query := `
		DELETE FROM message WHERE id = $1
	`
	_, err := r.db.NamedExec(query, id)
	return err
}

func (r LocalMessageRepository) GetMsgsAsPage(
	sen string,
	fil domain.Filter,
) ([]*domain.Message, *domain.Metadata, error) {
	query := `
		SELECT COUNT(*) OVER(), id, sender_id, receiver_id, body, sent_at, delivered_at, read_at, version
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
		args = []any{&TotalRows, &m.ID, &m.SenderID, &m.ReceiverID, &m.Body, &SentAt, &DeliveredAt, &ReadAt, &m.Version}
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

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func parseTime(t *string) (*time.Time, error) {
	if t == nil || *t == "" {
		return nil, nil
	}
	ti, err := time.Parse("2006-01-02 15:04:05-07:00", *t)
	if err != nil {
		ti, err = time.Parse("2006-01-02T15:04:05.999999-07:00", *t)
	}
	return &ti, err
}
