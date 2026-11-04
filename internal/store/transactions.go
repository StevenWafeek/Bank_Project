package store

import (
	"database/sql"
	"errors"
	"time"

	"miebank/internal/models"
)

// ApplyTransaction atomically changes a user's balance by delta (positive for deposit,
// negative for withdrawal) and records the transaction.
func (s *Store) ApplyTransaction(userID int64, txType string, delta int64) (models.Transaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.Transaction{}, err
	}
	defer tx.Rollback()

	var balance int64
	err = tx.QueryRow(`SELECT balance_cents FROM users WHERE id = ?`, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Transaction{}, ErrUserNotFound
	}
	if err != nil {
		return models.Transaction{}, err
	}
	newBalance := balance + delta
	if newBalance < 0 {
		return models.Transaction{}, ErrInsufficientFunds
	}
	if _, err := tx.Exec(`UPDATE users SET balance_cents = ? WHERE id = ?`, newBalance, userID); err != nil {
		return models.Transaction{}, err
	}

	amount := delta
	if amount < 0 {
		amount = -amount
	}
	t := models.Transaction{Type: txType, AmountCents: amount, BalanceCents: newBalance, CreatedAt: time.Now().UTC()}
	res, err := tx.Exec(
		`INSERT INTO transactions (user_id, type, amount_cents, balance_cents, created_at) VALUES (?, ?, ?, ?, ?)`,
		userID, t.Type, t.AmountCents, t.BalanceCents, t.CreatedAt,
	)
	if err != nil {
		return models.Transaction{}, err
	}
	if t.ID, err = res.LastInsertId(); err != nil {
		return models.Transaction{}, err
	}
	return t, tx.Commit()
}

// ListTransactions returns the user's most recent transactions, newest first.
func (s *Store) ListTransactions(userID int64, limit int) ([]models.Transaction, error) {
	rows, err := s.db.Query(
		`SELECT id, type, amount_cents, balance_cents, created_at FROM transactions
		 WHERE user_id = ? ORDER BY id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Transaction{}
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.Type, &t.AmountCents, &t.BalanceCents, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
