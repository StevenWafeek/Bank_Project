package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"miebank/internal/models"
)

// ApplyTransaction atomically changes a user's balance by delta (positive for deposit,
// negative for withdrawal) and records the transaction without description.
func (s *Store) ApplyTransaction(userID int64, txType string, delta int64) (models.Transaction, error) {
	return s.ApplyTransactionWithDesc(userID, txType, delta, "")
}

// ApplyTransactionWithDesc atomically changes a user's balance and records the transaction with a description.
func (s *Store) ApplyTransactionWithDesc(userID int64, txType string, delta int64, desc string) (models.Transaction, error) {
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
	t := models.Transaction{
		Type:         txType,
		AmountCents:  amount,
		BalanceCents: newBalance,
		Description:  desc,
		CreatedAt:    time.Now().UTC(),
	}
	res, err := tx.Exec(
		`INSERT INTO transactions (user_id, type, amount_cents, balance_cents, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, t.Type, t.AmountCents, t.BalanceCents, t.Description, t.CreatedAt,
	)
	if err != nil {
		return models.Transaction{}, err
	}
	if t.ID, err = res.LastInsertId(); err != nil {
		return models.Transaction{}, err
	}
	return t, tx.Commit()
}

// Transfer atomically transfers funds from sender to recipient.
func (s *Store) Transfer(senderID int64, recipientQuery string, amountCents int64, note string) (models.Transaction, models.User, error) {
	if amountCents <= 0 {
		return models.Transaction{}, models.User{}, errors.New("amount must be greater than 0")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return models.Transaction{}, models.User{}, err
	}
	defer tx.Rollback()

	// 1. Fetch sender
	var sender models.User
	err = tx.QueryRow(`SELECT id, name, email, password_hash, account_number, card_number, card_expiry, iban, balance_cents, created_at FROM users WHERE id = ?`, senderID).
		Scan(&sender.ID, &sender.Name, &sender.Email, &sender.PasswordHash, &sender.AccountNumber, &sender.CardNumber, &sender.CardExpiry, &sender.IBAN, &sender.BalanceCents, &sender.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Transaction{}, models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.Transaction{}, models.User{}, err
	}

	if sender.BalanceCents < amountCents {
		return models.Transaction{}, models.User{}, ErrInsufficientFunds
	}

	// 2. Fetch recipient (by account number, email, card number, or IBAN)
	trimmed := strings.TrimSpace(recipientQuery)
	upper := strings.ToUpper(trimmed)
	lower := strings.ToLower(trimmed)
	cleanCard := CleanCardNumber(trimmed)
	cleanIb := CleanIBAN(trimmed)

	var recipient models.User
	err = tx.QueryRow(
		`SELECT id, name, email, password_hash, account_number, card_number, card_expiry, iban, balance_cents, created_at
		 FROM users
		 WHERE account_number = ?
		    OR email = ?
		    OR (card_number != '' AND card_number = ?)
		    OR (iban != '' AND iban = ?)`,
		upper, lower, cleanCard, cleanIb,
	).Scan(&recipient.ID, &recipient.Name, &recipient.Email, &recipient.PasswordHash, &recipient.AccountNumber, &recipient.CardNumber, &recipient.CardExpiry, &recipient.IBAN, &recipient.BalanceCents, &recipient.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Transaction{}, models.User{}, ErrRecipientNotFound
	}
	if err != nil {
		return models.Transaction{}, models.User{}, err
	}

	if sender.ID == recipient.ID {
		return models.Transaction{}, models.User{}, ErrSelfTransfer
	}

	now := time.Now().UTC()
	senderNewBalance := sender.BalanceCents - amountCents
	recipientNewBalance := recipient.BalanceCents + amountCents

	// 3. Update sender balance
	if _, err := tx.Exec(`UPDATE users SET balance_cents = ? WHERE id = ?`, senderNewBalance, sender.ID); err != nil {
		return models.Transaction{}, models.User{}, err
	}

	// 4. Update recipient balance
	if _, err := tx.Exec(`UPDATE users SET balance_cents = ? WHERE id = ?`, recipientNewBalance, recipient.ID); err != nil {
		return models.Transaction{}, models.User{}, err
	}

	// Format notes
	memo := ""
	if trimmedNote := strings.TrimSpace(note); trimmedNote != "" {
		memo = fmt.Sprintf(" - \"%s\"", trimmedNote)
	}

	senderDesc := fmt.Sprintf("Transfer to %s (%s)%s", recipient.Name, recipient.AccountNumber, memo)
	recipientDesc := fmt.Sprintf("Transfer from %s (%s)%s", sender.Name, sender.AccountNumber, memo)

	// 5. Insert sender transaction
	senderTx := models.Transaction{
		Type:         models.TxTransferOut,
		AmountCents:  amountCents,
		BalanceCents: senderNewBalance,
		Description:  senderDesc,
		CreatedAt:    now,
	}
	res, err := tx.Exec(
		`INSERT INTO transactions (user_id, type, amount_cents, balance_cents, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		sender.ID, senderTx.Type, senderTx.AmountCents, senderTx.BalanceCents, senderTx.Description, senderTx.CreatedAt,
	)
	if err != nil {
		return models.Transaction{}, models.User{}, err
	}
	if senderTx.ID, err = res.LastInsertId(); err != nil {
		return models.Transaction{}, models.User{}, err
	}

	// 6. Insert recipient transaction
	if _, err := tx.Exec(
		`INSERT INTO transactions (user_id, type, amount_cents, balance_cents, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		recipient.ID, models.TxTransferIn, amountCents, recipientNewBalance, recipientDesc, now,
	); err != nil {
		return models.Transaction{}, models.User{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Transaction{}, models.User{}, err
	}

	return senderTx, recipient, nil
}

// PayBill debits the user and records a bill payment transaction.
func (s *Store) PayBill(userID int64, biller, reference string, amountCents int64) (models.Transaction, error) {
	desc := fmt.Sprintf("Bill Payment: %s (Ref: %s)", strings.TrimSpace(biller), strings.TrimSpace(reference))
	return s.ApplyTransactionWithDesc(userID, models.TxBillPayment, -amountCents, desc)
}

// Donate debits the user and records a charitable donation transaction.
func (s *Store) Donate(userID int64, charity, note string, amountCents int64) (models.Transaction, error) {
	desc := fmt.Sprintf("Donation to %s", strings.TrimSpace(charity))
	if trimmedNote := strings.TrimSpace(note); trimmedNote != "" {
		desc += fmt.Sprintf(" - \"%s\"", trimmedNote)
	}
	return s.ApplyTransactionWithDesc(userID, models.TxDonation, -amountCents, desc)
}

// ListTransactions returns the user's most recent transactions, newest first.
func (s *Store) ListTransactions(userID int64, limit int) ([]models.Transaction, error) {
	rows, err := s.db.Query(
		`SELECT id, type, amount_cents, balance_cents, description, created_at FROM transactions
		 WHERE user_id = ? ORDER BY id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Transaction{}
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.Type, &t.AmountCents, &t.BalanceCents, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
