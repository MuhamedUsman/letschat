package repository

import (
	"database/sql"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
)

type LocalMessageRepository struct {
	db *DB
}

func NewLocalMessageRepository(db *DB) LocalMessageRepository {
	return LocalMessageRepository{db}
}

type LatestMsgs map[string]*string

func (r LocalMessageRepository) GetLatestMsgBodyForConvos(cui ...string) (LatestMsgs, error) {
	// cui convoUsrId
	query := `
		SELECT body
		FROM message
		WHERE sender_id = $1 OR receiver_id = $1
	`
	msgs := make(LatestMsgs)
	for _, id := range cui {
		var msg string
		if err := r.db.QueryRowx(query, id).StructScan(&msg); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue // Ignore and continue, as there may be no messages for some chats
			}
			return nil, err
		}
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
		INSERT INTO message (id, sender_id, receiver_id, body, sent_at, delivered_at, read_at, version)
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
