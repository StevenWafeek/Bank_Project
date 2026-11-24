package store

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	name           TEXT    NOT NULL,
	email          TEXT    NOT NULL UNIQUE COLLATE NOCASE,
	password_hash  TEXT    NOT NULL,
	account_number TEXT    NOT NULL UNIQUE,
	card_number    TEXT    NOT NULL DEFAULT '',
	card_expiry    TEXT    NOT NULL DEFAULT '',
	iban           TEXT    NOT NULL DEFAULT '',
	balance_cents  INTEGER NOT NULL DEFAULT 0 CHECK (balance_cents >= 0),
	created_at     TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_card ON users(card_number) WHERE card_number != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_iban ON users(iban) WHERE iban != '';

CREATE TABLE IF NOT EXISTS transactions (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	type          TEXT    NOT NULL CHECK (type IN ('deposit', 'withdraw', 'transfer_out', 'transfer_in', 'donation', 'bill_payment')),
	amount_cents  INTEGER NOT NULL CHECK (amount_cents > 0),
	balance_cents INTEGER NOT NULL,
	description   TEXT    NOT NULL DEFAULT '',
	created_at    TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transactions_user ON transactions(user_id, id DESC);
`
