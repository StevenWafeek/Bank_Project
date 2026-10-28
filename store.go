package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrUserExists        = errors.New("an account with this email already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Transaction struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	AmountCents  int64     `json:"amount_cents"`
	BalanceCents int64     `json:"balance_cents"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	ID            int64
	Name          string
	Email         string
	PasswordHash  string
	AccountNumber string
	BalanceCents  int64
	CreatedAt     time.Time
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	name           TEXT    NOT NULL,
	email          TEXT    NOT NULL UNIQUE COLLATE NOCASE,
	password_hash  TEXT    NOT NULL,
	account_number TEXT    NOT NULL UNIQUE,
	balance_cents  INTEGER NOT NULL DEFAULT 0 CHECK (balance_cents >= 0),
	created_at     TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	type          TEXT    NOT NULL CHECK (type IN ('deposit', 'withdraw')),
	amount_cents  INTEGER NOT NULL CHECK (amount_cents > 0),
	balance_cents INTEGER NOT NULL,
	created_at    TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transactions_user ON transactions(user_id, id DESC);
`

// Store persists users and transactions in a SQLite database.
type Store struct {
	db *sql.DB
}

func NewStore(path string) (*Store, error) {
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

const userColumns = `id, name, email, password_hash, account_number, balance_cents, created_at`

func scanUser(row *sql.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.AccountNumber, &u.BalanceCents, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	return u, err
}

func (s *Store) CreateUser(name, email, passwordHash, accountNumber string) (User, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`, email).Scan(&exists); err != nil {
		return User{}, err
	}
	if exists {
		return User{}, ErrUserExists
	}
	now := time.Now().UTC()
	res, err := tx.Exec(
		`INSERT INTO users (name, email, password_hash, account_number, created_at) VALUES (?, ?, ?, ?, ?)`,
		name, email, passwordHash, accountNumber, now,
	)
	if err != nil {
		return User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	return User{ID: id, Name: name, Email: email, PasswordHash: passwordHash, AccountNumber: accountNumber, CreatedAt: now}, nil
}

func (s *Store) GetByEmail(email string) (User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE email = ?`, email))
}

func (s *Store) GetByID(id int64) (User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE id = ?`, id))
}

// ApplyTransaction atomically changes a user's balance by delta (positive for deposit,
// negative for withdrawal) and records the transaction.
func (s *Store) ApplyTransaction(userID int64, txType string, delta int64) (Transaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Transaction{}, err
	}
	defer tx.Rollback()

	var balance int64
	err = tx.QueryRow(`SELECT balance_cents FROM users WHERE id = ?`, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, ErrUserNotFound
	}
	if err != nil {
		return Transaction{}, err
	}
	newBalance := balance + delta
	if newBalance < 0 {
		return Transaction{}, ErrInsufficientFunds
	}
	if _, err := tx.Exec(`UPDATE users SET balance_cents = ? WHERE id = ?`, newBalance, userID); err != nil {
		return Transaction{}, err
	}

	amount := delta
	if amount < 0 {
		amount = -amount
	}
	t := Transaction{Type: txType, AmountCents: amount, BalanceCents: newBalance, CreatedAt: time.Now().UTC()}
	res, err := tx.Exec(
		`INSERT INTO transactions (user_id, type, amount_cents, balance_cents, created_at) VALUES (?, ?, ?, ?, ?)`,
		userID, t.Type, t.AmountCents, t.BalanceCents, t.CreatedAt,
	)
	if err != nil {
		return Transaction{}, err
	}
	if t.ID, err = res.LastInsertId(); err != nil {
		return Transaction{}, err
	}
	return t, tx.Commit()
}

// ListTransactions returns the user's most recent transactions, newest first.
func (s *Store) ListTransactions(userID int64, limit int) ([]Transaction, error) {
	rows, err := s.db.Query(
		`SELECT id, type, amount_cents, balance_cents, created_at FROM transactions
		 WHERE user_id = ? ORDER BY id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Transaction{}
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.Type, &t.AmountCents, &t.BalanceCents, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
