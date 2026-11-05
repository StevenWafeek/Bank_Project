// Package models holds the domain types shared across the backend.
package models

import "time"

const (
	TxDeposit     = "deposit"
	TxWithdraw    = "withdraw"
	TxTransferOut = "transfer_out"
	TxTransferIn  = "transfer_in"
	TxDonation    = "donation"
	TxBillPayment = "bill_payment"
)

type User struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	AccountNumber string    `json:"account_number"`
	BalanceCents  int64     `json:"balance_cents"`
	CreatedAt     time.Time `json:"created_at"`
}

type Transaction struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	AmountCents  int64     `json:"amount_cents"`
	BalanceCents int64     `json:"balance_cents"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
