// Package store persists users and transactions in SQLite.
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

var (
	ErrUserExists        = errors.New("an account with this email already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrSelfTransfer      = errors.New("cannot transfer to your own account")
	ErrRecipientNotFound = errors.New("recipient account not found")
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

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate schema: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return s, nil
}

// migrate ensures existing transactions table has description column and updated check constraint,
// and users table has card_number, card_expiry, and iban columns.
func (s *Store) migrate() error {
	var count int
	err := s.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='transactions'`).Scan(&count)
	if err == nil && count > 0 {
		var tblSQL string
		_ = s.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='transactions'`).Scan(&tblSQL)

		needsConstraintUpdate := !strings.Contains(tblSQL, "transfer_out")
		if needsConstraintUpdate {
			tx, err := s.db.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()

			if _, err := tx.Exec(`ALTER TABLE transactions RENAME TO transactions_old`); err != nil {
				return err
			}

			newTableSQL := `
			CREATE TABLE transactions (
				id            INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				type          TEXT    NOT NULL CHECK (type IN ('deposit', 'withdraw', 'transfer_out', 'transfer_in', 'donation', 'bill_payment')),
				amount_cents  INTEGER NOT NULL CHECK (amount_cents > 0),
				balance_cents INTEGER NOT NULL,
				description   TEXT    NOT NULL DEFAULT '',
				created_at    TIMESTAMP NOT NULL
			);`
			if _, err := tx.Exec(newTableSQL); err != nil {
				return err
			}

			hasDesc := strings.Contains(tblSQL, "description")
			var copySQL string
			if hasDesc {
				copySQL = `INSERT INTO transactions (id, user_id, type, amount_cents, balance_cents, description, created_at)
				           SELECT id, user_id, type, amount_cents, balance_cents, description, created_at FROM transactions_old`
			} else {
				copySQL = `INSERT INTO transactions (id, user_id, type, amount_cents, balance_cents, description, created_at)
				           SELECT id, user_id, type, amount_cents, balance_cents, '', created_at FROM transactions_old`
			}
			if _, err := tx.Exec(copySQL); err != nil {
				return err
			}
			if _, err := tx.Exec(`DROP TABLE transactions_old`); err != nil {
				return err
			}
			if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_user ON transactions(user_id, id DESC)`); err != nil {
				return err
			}

			if err := tx.Commit(); err != nil {
				return err
			}
		}
	}

	var userCount int
	_ = s.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='users'`).Scan(&userCount)
	if userCount > 0 {
		var userSQL string
		_ = s.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='users'`).Scan(&userSQL)
		if userSQL != "" {
			if !strings.Contains(userSQL, "card_number") {
				if _, err := s.db.Exec(`ALTER TABLE users ADD COLUMN card_number TEXT NOT NULL DEFAULT ''`); err != nil {
					return err
				}
			}
			if !strings.Contains(userSQL, "card_expiry") {
				if _, err := s.db.Exec(`ALTER TABLE users ADD COLUMN card_expiry TEXT NOT NULL DEFAULT ''`); err != nil {
					return err
				}
			}
			if !strings.Contains(userSQL, "iban") {
				if _, err := s.db.Exec(`ALTER TABLE users ADD COLUMN iban TEXT NOT NULL DEFAULT ''`); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Store) Close() error { return s.db.Close() }
