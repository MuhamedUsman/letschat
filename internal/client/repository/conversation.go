package repository

import (
	"database/sql"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
)

type LocalConversationRepository struct {
	db *DB
}

func NewLocalConversationRepository(db *DB) LocalConversationRepository {
	return LocalConversationRepository{db}
}

func (r LocalConversationRepository) SaveConversations(convos ...*domain.Conversation) error {
	query := `
		INSERT INTO conversation(user_id, username, user_email) 
		VALUES (:user_id, :username, :user_email)
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
		SELECT user_id, username, user_email 
		FROM conversation
		WHERE user_id = :user_id  
	`
	var c domain.Conversation
	if err := r.db.QueryRowx(query, id).StructScan(&c); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r LocalConversationRepository) GetConversations() ([]*domain.Conversation, error) {
	query := `
		SELECT user_id, username, user_email FROM conversation
	`
	rows, _ := r.db.Queryx(query)
	convos := make([]*domain.Conversation, 0)
	for rows.Next() {
		var c domain.Conversation
		if err := rows.StructScan(&c); err != nil {
			return nil, err
		}
		convos = append(convos, &c)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return convos, nil
}
