package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
)

var _ domain.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) RegisterUser(ctx context.Context, u *domain.User) (string, error) {
	// it takes less time around 2 ms if we check like this the approach below takes 400+ ms
	tx := contextGetTX(ctx)
	query := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id
		`
	args := []any{u.Name, u.Email, u.Password}
	var userID string
	var err error
	if tx != nil {
		err = tx.QueryRowxContext(ctx, query, args...).Scan(&userID)
	} else {
		err = r.db.QueryRowxContext(ctx, query, args...).Scan(&userID)
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" { // 400+ ms
				return "", domain.ErrDuplicateEmail
			}
		}
	}
	return userID, nil
}

func (r *UserRepository) ExistsUser(ctx context.Context, email string) (bool, error) {
	tx := contextGetTX(ctx)
	var err error
	existsQuery := `SELECT EXISTS(SELECT TRUE FROM users WHERE email = $1)`
	exists := false
	if tx != nil {
		err = tx.QueryRowContext(ctx, existsQuery, email).Scan(&exists)
	} else {
		err = r.db.QueryRowContext(ctx, existsQuery, email).Scan(&exists)
	}
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) GetByUniqueField(ctx context.Context, fieldName, fieldValue string) (*domain.User, error) {
	query := `
		SELECT * 
		FROM users
		WHERE %v = $1
		`
	query = fmt.Sprintf(query, fieldName)
	tx := contextGetTX(ctx)
	var user domain.User
	var err error
	if tx != nil {
		err = tx.QueryRowxContext(ctx, query, fieldValue).StructScan(&user)
	} else {
		err = r.db.QueryRowxContext(ctx, query, fieldValue).StructScan(&user)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users 
		SET name = :name, email = :email, password = :password, version = version + 1
		WHERE id = :id AND version = :version
		`
	tx := contextGetTX(ctx)
	var err error
	var editStatus sql.Result
	if tx != nil {
		editStatus, err = tx.NamedExecContext(ctx, query, u)
	} else {
		editStatus, err = r.db.NamedExecContext(ctx, query, u)
	}
	if err != nil {
		return err
	}
	rowsAffected, err := editStatus.RowsAffected()
	if err != nil {
		slog.Error(err.Error())
	}
	if rowsAffected == 0 {
		return domain.ErrEditConflict
	}
	return nil
}

func (r *UserRepository) GetForToken(ctx context.Context, scope string, hash []byte) (*domain.User, error) {
	query := `
		SELECT * FROM users u
	    WHERE id IN (
	    SELECT user_id
	    FROM token WHERE scope = $1 
         AND hash = $2 
         AND expiry > NOW())
		`
	var usr domain.User
	var err error
	if tx := contextGetTX(ctx); tx != nil {
		err = tx.QueryRowxContext(ctx, query, scope, hash).StructScan(&usr)
	} else {
		err = r.db.QueryRowxContext(ctx, query, scope, hash).StructScan(&usr)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return &usr, nil
}

func (r *UserRepository) ActivateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET activated = TRUE, version = version + 1
		WHERE id = :id AND version = :version
		`
	var result sql.Result
	var err error
	if tx := contextGetTX(ctx); tx != nil {
		result, err = tx.NamedExecContext(ctx, query, user)
	} else {
		result, err = r.db.NamedExecContext(ctx, query, user)
	}
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrEditConflict
	}
	return nil
}
