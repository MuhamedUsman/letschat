package repository

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/jmoiron/sqlx"
)

var _ domain.MessageRepository = (*MessageRepository)(nil)

type MessageRepository struct {
	db *DB
}

func NewMessageRepository(db *DB) *MessageRepository {
	return &MessageRepository{db}
}

func (r *MessageRepository) GetByID(ctx context.Context, id string, op domain.MsgOperation) (*domain.Message, error) {
	query := `
		SELECT * FROM message 
        WHERE id = $1
		AND operation = $2
        `
	var message domain.Message
	var err error
	if tx := contextGetTX(ctx); tx != nil {
		err = tx.QueryRowxContext(ctx, query, id, op).StructScan(&message)
	} else {
		err = r.db.QueryRowxContext(ctx, query, id, op).StructScan(&message)
	}
	return &message, err
}

func (r *MessageRepository) GetUnDeliveredMessages(ctx context.Context, rcvrID string, op domain.MsgOperation, c domain.MsgChan) error {
	query := `
		SELECT *
		FROM message
		WHERE receiver_id = $1 AND operation = $2
		ORDER BY sent_at
		`
	var rows *sqlx.Rows
	if tx := contextGetTX(ctx); tx != nil {
		rows, _ = tx.QueryxContext(ctx, query, rcvrID, op)
	} else {
		rows, _ = r.db.QueryxContext(ctx, query, rcvrID, op)
	}
	defer rows.Close()
	for rows.Next() {
		var msg domain.Message
		if err := rows.StructScan(&msg); err != nil {
			return err
		}
		c <- &msg
	}
	if err := rows.Err(); err != nil {
		return rows.Err()
	}
	return nil
}

func (r *MessageRepository) GetMessagesAsPage(
	ctx context.Context,
	rcvrID string,
	c domain.MsgChan,
	filter *domain.Filter,
) (*domain.Metadata, error) {
	query := `
		SELECT COUNT(*) OVER() total_rows, *
		FROM message
		WHERE receiver_id = $1
		ORDER BY sent_at DESC
		LIMIT $2
	    OFFSET $3
		`
	var rows *sqlx.Rows
	args := []any{rcvrID, filter.Limit(), filter.Offset()}
	if tx := contextGetTX(ctx); tx != nil {
		rows, _ = tx.QueryxContext(ctx, query, args...)
	} else {
		rows, _ = r.db.QueryxContext(ctx, query, args...)
	}
	var totalRows int
	for rows.Next() {
		var row struct {
			domain.Message
			TotalRows int `db:"total_rows"`
		}
		if err := rows.StructScan(&row); err != nil {
			return &domain.Metadata{}, err
		}
		c <- &row.Message
		totalRows = row.TotalRows
	}
	metadata := domain.CalculateMetadata(totalRows, filter.PageSize, filter.Page)
	return &metadata, nil
}

func (r *MessageRepository) InsertMessage(ctx context.Context, m *domain.Message) error {
	query := `
		INSERT INTO message (id, sender_id, receiver_id, body, sent_at, delivered_at, read_at, operation) 
		VALUES (:id, :sender_id, :receiver_id, :body, :sent_at, :delivered_at, :read_at, :operation)
		ON CONFLICT (id)
		DO UPDATE SET
		              sender_id = EXCLUDED.sender_id,
		              receiver_id = EXCLUDED.receiver_id,
		              body = EXCLUDED.body,
		              sent_at = EXCLUDED.sent_at,
		              delivered_at = EXCLUDED.delivered_at,
		              read_at = EXCLUDED.read_at,
		              operation = EXCLUDED.operation
		`
	if tx := contextGetTX(ctx); tx != nil {
		_, err := tx.NamedExecContext(ctx, query, m)
		return err
	}
	_, err := r.db.NamedExecContext(ctx, query, m)
	return err
}

func (r *MessageRepository) DeleteMessageWithOperation(ctx context.Context, mID string, op domain.MsgOperation) error {
	query := `
		DELETE FROM message 
        WHERE id = $1 AND operation = $2
        `
	if tx := contextGetTX(ctx); tx != nil {
		_, err := tx.ExecContext(ctx, query, mID, op)
		return err
	}
	_, err := r.db.ExecContext(ctx, query, mID)
	return err
}

func (r *MessageRepository) DeleteMessage(ctx context.Context, mID string) error {
	query := `
		DELETE FROM message 
        WHERE id = $1
        `
	if tx := contextGetTX(ctx); tx != nil {
		_, err := tx.ExecContext(ctx, query, mID)
		return err
	}
	_, err := r.db.ExecContext(ctx, query, mID)
	return err
}
