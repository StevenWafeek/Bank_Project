// Package money validates deposit and withdrawal amounts.
package money

import (
	"errors"
	"math"
)

const (
	MinTxAmountCents = 20_00        // $20.00 minimum per transaction
	MaxTxAmountCents = 1_000_000_00 // $1,000,000.00 maximum per transaction
)

// ParseAmount converts a whole-dollar amount to cents, rejecting decimals and out-of-range values.
func ParseAmount(amount float64) (int64, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return 0, errors.New("amount must be greater than 0")
	}
	if amount != math.Trunc(amount) {
		return 0, errors.New("amount must be a whole number (no decimals)")
	}
	cents := amount * 100
	if cents < MinTxAmountCents {
		return 0, errors.New("the minimum amount for a deposit or withdrawal is $20.00")
	}
	if cents > MaxTxAmountCents {
		return 0, errors.New("amount exceeds the per-transaction limit of 1,000,000.00")
	}
	return int64(cents), nil
}
