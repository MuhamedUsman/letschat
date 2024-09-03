package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/jmoiron/sqlx"
	"strings"
)

var _ domain.MessageRepository = (*MessageRepository)(nil)

type MessageRepository struct {
	db *DB
}

func NewMessageRepository(db *DB) *MessageRepository {
	return &MessageRepository{db}
}

func (r *MessageRepository) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	query := `
		SELECT * FROM message 
        WHERE id = $1
        `
	var message domain.Message
	var err error
	if tx := contextGetTX(ctx); tx != nil {
		err = tx.QueryRowxContext(ctx, query, id).StructScan(&message)
	} else {
		err = r.db.QueryRowxContext(ctx, query, id).StructScan(&message)
	}
	return &message, err
}

func (r *MessageRepository) GetUnreadMessages(ctx context.Context, rcvrID string, c domain.MsgChan) error {
	query := `
		SELECT *
		FROM message
		WHERE receiver_id = $1 AND read_at IS NULL
		ORDER BY sent_at
		`
	var rows *sqlx.Rows
	if tx := contextGetTX(ctx); tx != nil {
		rows, _ = tx.QueryxContext(ctx, query, rcvrID)
	} else {
		rows, _ = r.db.QueryxContext(ctx, query, rcvrID)
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
		INSERT INTO message (id, sender_id, receiver_id, body, sent_at) 
		VALUES (:id, :sender_id, :receiver_id, :body, :sent_at)
		`
	if tx := contextGetTX(ctx); tx != nil {
		_, err := tx.NamedExecContext(ctx, query, m)
		return err
	}
	_, err := r.db.NamedExecContext(ctx, query, m)
	return err
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, m *domain.Message) error {
	var setStatements []string
	args := map[string]any{
		"id":      m.ID,
		"version": m.Version,
	}
	if m.DeliveredAt != nil {
		setStatements = append(setStatements, "delivered_at = :delivered_at")
		args["delivered_at"] = m.DeliveredAt
	}
	if m.ReadAt != nil {
		setStatements = append(setStatements, "read_at = :read_at")
		args["read_at"] = m.ReadAt
	}
	setStatements = append(setStatements, "version = version + 1")
	if len(setStatements) == 1 {
		return nil
	}
	query := fmt.Sprintf(`
		UPDATE message 
		SET %s 
		WHERE id = :id AND version = :version
	`, strings.Join(setStatements, ", "))
	tx := contextGetTX(ctx)
	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.NamedExecContext(ctx, query, args)
	} else {
		result, err = r.db.NamedExecContext(ctx, query, args)
	}
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrEditConflict
	}
	return nil
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
