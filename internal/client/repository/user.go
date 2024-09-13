package repository

import (
	"database/sql"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
)

type LocalUserRepository struct {
	db *DB
}

func newLocalUserRepository(db *DB) LocalUserRepository {
	return LocalUserRepository{db}
}

func (r LocalUserRepository) GetCurrentUser() (*domain.User, error) {
	query := `
		SELECT * FROM users
	`
	var usr domain.User
	if err := r.db.QueryRowx(query).StructScan(&usr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return &usr, nil
}

func (r LocalUserRepository) SaveCurrentUser(u *domain.User) error {
	query := `
		INSERT INTO users (id, name, email, created_at) 
		VALUES (:id, :name, :email, :created_at)
	`
	_, err := r.db.NamedExec(query, u)
	return err
}

func (r LocalUserRepository) DeletePreviousUser() error {
	query := `
		DELETE FROM users
	`
	_, err := r.db.Exec(query)
	return err
}
