package repository

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/api/utility"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"time"
)

type ctxKey string

const txCtxKey = ctxKey("USER")

func contextGetTX(ctx context.Context) *TX {
	tx, ok := ctx.Value(txCtxKey).(*TX)
	if !ok {
		return nil
	}
	return tx
}

type DB struct {
	*sqlx.DB
}

func OpenDB(cfg *utility.Config) *DB {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := sqlx.ConnectContext(ctx, "pgx", cfg.DB.DSN)
	if err != nil {
		panic("failed to connect database")
	}
	idleDuration, err := time.ParseDuration(cfg.DB.MaxIdleConnTime)
	if err != nil {
		panic("failed to parse max idle connection time, valid defaults must be set")
	}
	db.SetMaxOpenConns(cfg.DB.MaxOpenConn)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConn)
	db.SetConnMaxIdleTime(idleDuration)
	return &DB{db}
}

type TX struct {
	*sqlx.Tx
}

func (db *DB) BeginTx(ctx context.Context) (*TX, error) {
	txx, err := db.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &TX{txx}, err
}

func (db *DB) RunInTX(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	ctx = context.WithValue(ctx, txCtxKey, tx)
	if err = fn(ctx); err != nil {
		return err
	}
	return tx.Commit()
}
