package repository

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/domain"
)

var _ domain.TokenRepository = (*TokenRepository)(nil)

type TokenRepository struct {
	db *DB
}

func NewTokenRepository(db *DB) *TokenRepository {
	return &TokenRepository{db}
}

func (r *TokenRepository) Insert(ctx context.Context, token *domain.Token) error {
	query := `
		INSERT INTO token (hash, user_id, expiry, scope) 
		VALUES (:hash, :user_id, :expiry, :scope)
		`
	tx := contextGetTX(ctx)
	var err error
	if tx != nil {
		_, err = tx.NamedExecContext(ctx, query, token)
	} else {
		_, err = r.db.NamedExecContext(ctx, query, token)
	}
	return err
}

func (r *TokenRepository) DeleteAllForUser(ctx context.Context, userID, scope string) error {
	query := `
		DELETE FROM token 
        WHERE user_id = $1 AND scope = $2
        `
	tx := contextGetTX(ctx)
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, userID, scope)
	} else {
		_, err = r.db.ExecContext(ctx, query, userID, scope)
	}
	return err
}
