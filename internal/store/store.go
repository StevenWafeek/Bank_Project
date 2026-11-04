// Package store persists users and transactions in SQLite.
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var (
	ErrUserExists        = errors.New("an account with this email already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Store struct {
	db *sql.DB
}

// New opens (or creates) the SQLite database at path and applies the schema.
func New(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite allows a single writer; one connection avoids "database is locked" errors.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }
