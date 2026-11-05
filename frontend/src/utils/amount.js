import { formatMoney } from './format.js';

export const MIN_AMOUNT = 20;
export const MAX_AMOUNT = 1_000_000;
export const QUICK_AMOUNTS = [20, 50, 100, 500];

// Keys that would let users type decimals, exponents or signs into the amount field.
export const BLOCKED_KEYS = ['.', ',', 'e', 'E', '+', '-'];

// Returns an error message for an invalid amount, or null when it is valid.
export function validateAmount(raw) {
  const amount = Number(raw);
  if (raw === '' || raw == null || !Number.isFinite(amount) || amount <= 0) {
    return 'Please enter a valid amount greater than 0.';
  }
  if (!Number.isInteger(amount)) return 'Please enter a whole number (no decimals).';
  if (amount < MIN_AMOUNT) {
    return `The minimum amount for a deposit or withdrawal is ${formatMoney(MIN_AMOUNT * 100)}.`;
  }
  if (amount > MAX_AMOUNT) return 'Amount exceeds the per-transaction limit of 1,000,000.00';
  return null;
}

// Validates whole-dollar amounts for transfers, bills, or donations ($1.00 min).
export function validatePaymentAmount(raw, min = 1) {
  const amount = Number(raw);
  if (raw === '' || raw == null || !Number.isFinite(amount) || amount <= 0) {
    return 'Please enter a valid amount greater than 0.';
  }
  if (!Number.isInteger(amount)) return 'Please enter a whole number (no decimals).';
  if (amount < min) {
    return `The minimum amount is ${formatMoney(min * 100)}.`;
  }
  if (amount > MAX_AMOUNT) return 'Amount exceeds the per-transaction limit of 1,000,000.00';
  return null;
}
