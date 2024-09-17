package repository

import "github.com/M0hammadUsman/letschat/internal/domain"

type LocalConversationRepository struct {
	db *DB
}

func NewLocalConversationRepository(db *DB) LocalConversationRepository {
	return LocalConversationRepository{db}
}

func (r LocalConversationRepository) SaveConversations(convos ...*domain.Conversation) error {
	query := `
		INSERT INTO conversation(user_id, username, user_email, latest_msg) 
		VALUES (:user_id, :username, :user_email, :latest_msg, :last_online)
	`
	for _, convo := range convos {
		_, err := r.db.NamedExec(query, convo)
		if err != nil {
			return err
		}
	}
	return nil
}
func (r LocalConversationRepository) GetConversations() ([]*domain.Conversation, error) {
	query := `
		SELECT * FROM conversation
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
