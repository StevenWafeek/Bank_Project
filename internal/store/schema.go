package store

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
