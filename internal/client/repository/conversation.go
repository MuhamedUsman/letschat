package repository

import (
	"database/sql"
	"errors"
	"github.com/MuhamedUsman/letschat/internal/domain"
	"time"
)

type LocalConversationRepository struct {
	db *DB
}

func NewLocalConversationRepository(db *DB) LocalConversationRepository {
	return LocalConversationRepository{db}
}

func (r LocalConversationRepository) SaveConversations(convos ...*domain.Conversation) error {
	query := `
		INSERT INTO conversation(user_id, username, user_email, last_online) 
		VALUES (:user_id, :username, :user_email, :last_online)
	`
	for _, convo := range convos {
		_, err := r.db.NamedExec(query, convo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r LocalConversationRepository) DeleteAllConversations() error {
	query := `
		DELETE FROM conversation
	`
	_, err := r.db.Exec(query)
	return err
}

func (r LocalConversationRepository) GetConversationByUserID(id string) (*domain.Conversation, error) {
	query := `
		SELECT user_id, username, user_email, last_online
		FROM conversation
		WHERE user_id = :user_id  
	`
	var c domain.Conversation
	var LastOnline any
	args := []any{&c.UserID, &c.Username, &c.UserEmail, &LastOnline}
	if err := r.db.QueryRow(query, id).Scan(args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	if LastOnline != nil {
		if timeStr, ok := LastOnline.(string); ok {
			c.LastOnline, _ = parseTime(&timeStr)
		}
	}
	return &c, nil
}

func (r LocalConversationRepository) GetConversations() ([]*domain.Conversation, error) {
	query := `
		SELECT user_id, username, user_email, last_online FROM conversation
	`
	rows, _ := r.db.Queryx(query)
	convos := make([]*domain.Conversation, 0)
	for rows.Next() {
		var c domain.Conversation
		var LastOnline any
		args := []any{&c.UserID, &c.Username, &c.UserEmail, &LastOnline}
		if err := rows.Scan(args...); err != nil {
			return nil, err
		}
		if LastOnline != nil {
			if timeStr, ok := LastOnline.(time.Time); ok {
				c.LastOnline = &timeStr
			}
		}

		convos = append(convos, &c)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return convos, nil
}
