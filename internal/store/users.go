package store

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"miebank/internal/models"
)

const userColumns = `id, name, email, password_hash, account_number, balance_cents, created_at`

func scanUser(row *sql.Row) (models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.AccountNumber, &u.BalanceCents, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrUserNotFound
	}
	return u, err
}

func (s *Store) CreateUser(name, email, passwordHash, accountNumber string) (models.User, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.User{}, err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`, email).Scan(&exists); err != nil {
		return models.User{}, err
	}
	if exists {
		return models.User{}, ErrUserExists
	}
	now := time.Now().UTC()
	res, err := tx.Exec(
		`INSERT INTO users (name, email, password_hash, account_number, created_at) VALUES (?, ?, ?, ?, ?)`,
		name, email, passwordHash, accountNumber, now,
	)
	if err != nil {
		return models.User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.User{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.User{}, err
	}
	return models.User{ID: id, Name: name, Email: email, PasswordHash: passwordHash, AccountNumber: accountNumber, CreatedAt: now}, nil
}

func (s *Store) GetByEmail(email string) (models.User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE email = ?`, email))
}

func (s *Store) GetByID(id int64) (models.User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE id = ?`, id))
}

func (s *Store) GetByAccountNumber(accountNumber string) (models.User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE account_number = ?`, strings.ToUpper(strings.TrimSpace(accountNumber))))
}

func (s *Store) GetByEmailOrAccount(query string) (models.User, error) {
	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)
	lower := strings.ToLower(trimmed)
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE account_number = ? OR email = ?`, upper, lower))
}

func (s *Store) UpdateUserName(id int64, newName string) error {
	res, err := s.db.Exec(`UPDATE users SET name = ? WHERE id = ?`, strings.TrimSpace(newName), id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *Store) UpdateUserPassword(id int64, newPasswordHash string) error {
	res, err := s.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, newPasswordHash, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}
