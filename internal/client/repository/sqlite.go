package repository

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"time"
)

const ( // Local Database tables for client side application

	createUsersTable = `
		-- Just to store the current logged-in user
		CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL,
            created_at DATETIME NOT NULL -- DATETIME works TEXT, INTEGER will not be mapped to time.Time
		);
	`
	createMessageTable = `
		CREATE TABLE IF NOT EXISTS message (
            id TEXT PRIMARY KEY, -- Simulating UUID
            sender_id TEXT,
            receiver_id TEXT,
            body TEXT NOT NULL,
            sent_at TEXT,
            delivered_at DATETIME,
            read_at DATETIME,
            version INTEGER NOT NULL DEFAULT 1
		);
		CREATE INDEX IF NOT EXISTS idx_message_sender_receiver_sent_at ON message(sender_id, receiver_id, sent_at DESC);
	`
	createConversationTable = `
		CREATE TABLE IF NOT EXISTS conversation (
            user_id TEXT NOT NULL,
            username TEXT NOT NULL,
            user_email TEXT NOT NULL
		);
	`
)

type DB struct {
	*sqlx.DB
}

/*func OpenDB(filesDir string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := sqlx.ConnectContext(ctx, "sqlite3", filepath.Join(filesDir, "Letschat.db"))
	if err == nil {
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(5)
		db.SetConnMaxIdleTime(15 * time.Minute)
	}
	if err != nil && db != nil {
		db.Close()
	}
	return &DB{db}, err
}*/

func OpenDB(filesDir string, key int) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	name := fmt.Sprintf("Letschat%v.db", key)
	db, err := sqlx.ConnectContext(ctx, "sqlite3", filepath.Join(filesDir, name))
	if err == nil {
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(5)
		db.SetConnMaxIdleTime(15 * time.Minute)
	}
	if err != nil && db != nil {
		db.Close()
	}
	return &DB{db}, err
}

func DeleteDBFile(filesDir string) error {
	return os.Remove(filepath.Join(filesDir, "Letschat.db"))
}

func (db *DB) RunMigrations() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, createUsersTable); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, createMessageTable); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, createConversationTable); err != nil {
		return err
	}
	return nil
}
