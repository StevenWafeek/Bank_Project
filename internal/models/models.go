// Package models holds the domain types shared across the backend.
package models

import "time"

const (
	TxDeposit  = "deposit"
	TxWithdraw = "withdraw"
)

type User struct {
	ID            int64
	Name          string
	Email         string
	PasswordHash  string
	AccountNumber string
	BalanceCents  int64
	CreatedAt     time.Time
}

type Transaction struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	AmountCents  int64     `json:"amount_cents"`
	BalanceCents int64     `json:"balance_cents"`
	CreatedAt    time.Time `json:"created_at"`
}
