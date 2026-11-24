package store

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"miebank/internal/models"
)

const userColumns = `id, name, email, password_hash, account_number, card_number, card_expiry, iban, balance_cents, created_at`

func scanUser(row *sql.Row) (models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.AccountNumber, &u.CardNumber, &u.CardExpiry, &u.IBAN, &u.BalanceCents, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrUserNotFound
	}
	return u, err
}

// GenerateCardNumber returns a 16-digit debit card number starting with 5342 (MIeBank Debit).
func GenerateCardNumber() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("5342%012d", n.Int64()), nil
}

// GenerateCardExpiry returns MM/YY 5 years from given time.
func GenerateCardExpiry(t time.Time) string {
	exp := t.AddDate(5, 0, 0)
	return fmt.Sprintf("%02d/%02d", exp.Month(), exp.Year()%100)
}

// GenerateIBAN returns an IBAN format EG96MIEB + 16 numeric digits derived from account number.
func GenerateIBAN(accountNumber string) string {
	digits := strings.TrimPrefix(accountNumber, "MIE")
	var numOnly strings.Builder
	for _, ch := range digits {
		if ch >= '0' && ch <= '9' {
			numOnly.WriteRune(ch)
		}
	}
	res := numOnly.String()
	if len(res) > 16 {
		res = res[len(res)-16:]
	}
	padded := fmt.Sprintf("%016s", res)
	return fmt.Sprintf("EG96MIEB%s", padded)
}

// CleanCardNumber removes all spaces, hyphens, and non-digits from a card number string.
func CleanCardNumber(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// CleanIBAN removes spaces and hyphens and normalizes to uppercase.
func CleanIBAN(s string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(strings.TrimSpace(s)) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (s *Store) CreateUser(name, email, passwordHash, accountNumber string, cardDetails ...string) (models.User, error) {
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

	var cardNumber, cardExpiry, iban string
	if len(cardDetails) >= 3 {
		cardNumber = CleanCardNumber(cardDetails[0])
		cardExpiry = strings.TrimSpace(cardDetails[1])
		iban = CleanIBAN(cardDetails[2])
	}

	res, err := tx.Exec(
		`INSERT INTO users (name, email, password_hash, account_number, card_number, card_expiry, iban, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		name, email, passwordHash, accountNumber, cardNumber, cardExpiry, iban, now,
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
	return models.User{
		ID:            id,
		Name:          name,
		Email:         email,
		PasswordHash:  passwordHash,
		AccountNumber: accountNumber,
		CardNumber:    cardNumber,
		CardExpiry:    cardExpiry,
		IBAN:          iban,
		CreatedAt:     now,
	}, nil
}

func (s *Store) GetByEmail(email string) (models.User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE email = ?`, strings.ToLower(strings.TrimSpace(email))))
}

func (s *Store) GetByID(id int64) (models.User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE id = ?`, id))
}

func (s *Store) GetByAccountNumber(accountNumber string) (models.User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE account_number = ?`, strings.ToUpper(strings.TrimSpace(accountNumber))))
}

func (s *Store) GetByCardNumber(cardNum string) (models.User, error) {
	clean := CleanCardNumber(cardNum)
	if clean == "" {
		return models.User{}, ErrUserNotFound
	}
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE card_number = ?`, clean))
}

func (s *Store) GetByIBAN(iban string) (models.User, error) {
	clean := CleanIBAN(iban)
	if clean == "" {
		return models.User{}, ErrUserNotFound
	}
	return scanUser(s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE iban = ?`, clean))
}

func (s *Store) GetByEmailOrAccount(query string) (models.User, error) {
	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)
	lower := strings.ToLower(trimmed)
	cleanCard := CleanCardNumber(trimmed)
	cleanIb := CleanIBAN(trimmed)
	return scanUser(s.db.QueryRow(`
		SELECT `+userColumns+` FROM users
		WHERE account_number = ?
		   OR email = ?
		   OR (card_number != '' AND card_number = ?)
		   OR (iban != '' AND iban = ?)`,
		upper, lower, cleanCard, cleanIb))
}

func (s *Store) GetByLoginIdentifier(query string) (models.User, error) {
	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)
	lower := strings.ToLower(trimmed)
	cleanCard := CleanCardNumber(trimmed)
	cleanIb := CleanIBAN(trimmed)
	return scanUser(s.db.QueryRow(`
		SELECT `+userColumns+` FROM users
		WHERE email = ?
		   OR account_number = ?
		   OR (card_number != '' AND card_number = ?)
		   OR (iban != '' AND iban = ?)`,
		lower, upper, cleanCard, cleanIb))
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

func (s *Store) UpdateUserProfile(id int64, name, cardNumber, cardExpiry, iban string) error {
	name = strings.TrimSpace(name)
	cardNum := CleanCardNumber(cardNumber)
	cardExp := strings.TrimSpace(cardExpiry)
	ib := CleanIBAN(iban)

	res, err := s.db.Exec(`
		UPDATE users
		SET name = ?, card_number = ?, card_expiry = ?, iban = ?
		WHERE id = ?`,
		name, cardNum, cardExp, ib, id,
	)
	if err != nil {
		if strings.Contains(err.Error(), "idx_users_card") || strings.Contains(err.Error(), "card_number") {
			return errors.New("this card number is already associated with another account")
		}
		if strings.Contains(err.Error(), "idx_users_iban") || strings.Contains(err.Error(), "iban") {
			return errors.New("this IBAN is already associated with another account")
		}
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

func (s *Store) UpdateUserCardNumber(id int64, cardNumber string) error {
	clean := CleanCardNumber(cardNumber)
	res, err := s.db.Exec(`UPDATE users SET card_number = ? WHERE id = ?`, clean, id)
	if err != nil {
		if strings.Contains(err.Error(), "idx_users_card") || strings.Contains(err.Error(), "card_number") {
			return errors.New("this card number is already associated with another account")
		}
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

func (s *Store) UpdateUserCardExpiry(id int64, cardExpiry string) error {
	trimmed := strings.TrimSpace(cardExpiry)
	res, err := s.db.Exec(`UPDATE users SET card_expiry = ? WHERE id = ?`, trimmed, id)
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

func (s *Store) UpdateUserIBAN(id int64, iban string) error {
	clean := CleanIBAN(iban)
	res, err := s.db.Exec(`UPDATE users SET iban = ? WHERE id = ?`, clean, id)
	if err != nil {
		if strings.Contains(err.Error(), "idx_users_iban") || strings.Contains(err.Error(), "iban") {
			return errors.New("this IBAN is already associated with another account")
		}
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
