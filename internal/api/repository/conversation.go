package repository

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/jmoiron/sqlx"
)

var _ domain.ConversationRepository = (*ConversationRepository)(nil)

type ConversationRepository struct {
	DB *DB
}

func NewConversationRepository(db *DB) *ConversationRepository {
	return &ConversationRepository{DB: db}
}

func (r *ConversationRepository) CreateConversation(ctx context.Context, senderID, receiverID string) error {
	query := `
		INSERT INTO conversation 
		VALUES ($1, $2)
		`
	var err error
	if tx := contextGetTX(ctx); tx != nil {
		_, err = tx.Exec(query, senderID, receiverID)
	} else {
		_, err = r.DB.ExecContext(ctx, query, senderID, receiverID)
	}
	return err
}

func (r *ConversationRepository) GetConversations(ctx context.Context, usrID string) ([]*domain.Conversation, error) {
	/*query := `
			SELECT sender_id, sender.name sender_name, sender.email sender_email,
	            receiver_id, receiver.name receiver_name, receiver.email receiver_email,
	            CASE
	                WHEN sender_id = $1 THEN sender.last_online
	                ELSE receiver.last_online
	                END AS last_online
			FROM conversation
			    INNER JOIN users sender ON sender_id = sender.id
			    INNER JOIN users receiver ON receiver_id = receiver.id
			WHERE sender_id = $1 OR receiver_id = $1
			`*/
	query := `
		SELECT 
		    CASE 
		        WHEN sender_id = $1 THEN receiver_id
		        ELSE sender_id
	        END AS user_id,
	        CASE 
	            WHEN sender_id = $1 THEN receiver.name
	            ELSE sender.name
	        END AS username,
	        CASE 
	            WHEN sender_id = $1 THEN receiver.email
	            ELSE sender.email
	        END AS user_email,
	        CASE 
	            WHEN sender_id = $1 THEN receiver.last_online
	            ELSE sender.last_online
	        END AS last_online
		FROM conversation
		    INNER JOIN users sender ON sender_id = sender.id
		    INNER JOIN users receiver ON receiver_id = receiver.id
		WHERE sender_id = $1 OR receiver_id = $1
		`
	var rows *sqlx.Rows
	if tx := contextGetTX(ctx); tx != nil {
		rows, _ = tx.QueryxContext(ctx, query, usrID)
	} else {
		rows, _ = r.DB.QueryxContext(ctx, query, usrID)
	}
	defer rows.Close()
	conversations := make([]*domain.Conversation, 0)
	for rows.Next() {
		var c domain.Conversation
		_ = rows.StructScan(&c)
		conversations = append(conversations, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *ConversationRepository) ConversationExists(ctx context.Context, senderID, receiverID string) (bool, error) {
	query := `
		SELECT CASE
	    WHEN EXISTS (
		    SELECT FROM conversation 
	        WHERE sender_id = $1 AND receiver_id = $1
        )
	    THEN TRUE 
	    ELSE FALSE 
	    END
        `
	var exists bool
	var err error
	if tx := contextGetTX(ctx); tx != nil {
		err = tx.QueryRowContext(ctx, query, senderID, receiverID).Scan(&exists)
	} else {
		err = r.DB.QueryRowContext(ctx, query, senderID, receiverID).Scan(&exists)
	}
	if err != nil {
		return false, err
	}
	return exists, nil
}
