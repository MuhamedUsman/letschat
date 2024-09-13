package repository

import "github.com/M0hammadUsman/letschat/internal/domain"

type LocalConversationRepository struct {
	db *DB
}

func NewLocalConversationRepository(db *DB) LocalConversationRepository {
	return LocalConversationRepository{db}
}

func (c LocalConversationRepository) SaveConversations(convos ...*domain.Conversation) error {
	query := `
		INSERT INTO conversations VALUES (:userID, :username, :userEmail, :latestMsg)
	`
	_, err := c.db.NamedExec(query, convos)
	return err
}
