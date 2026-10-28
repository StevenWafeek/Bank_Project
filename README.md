# MIeBank

A simple banking web app: **Go** backend + vanilla **HTML/CSS/JS** frontend.

## Features
- Sign up / log in (bcrypt-hashed passwords, HTTP-only session cookies, 30-min sliding sessions)
- Dashboard with 3 options: **Your Balance**, **Withdraw**, **Deposit**
- Deposits and withdrawals must be whole dollar amounts (no decimals), at least $20 (max $1,000,000 per transaction)
- Recent transaction history
- About page for MIeBank

## Run
Requires Go 1.22+.

```bash
go run .
# open http://localhost:8080
```

Environment variables (optional):
| Var | Default |
| --- | --- |
| `ADDR` | `:8080` |
| `DB_PATH` | `data/miebank.db` |
| `STATIC_DIR` | `static` |

Data is stored in **SQLite** (pure-Go driver `modernc.org/sqlite`, no CGO/C compiler needed). Tables `users` and `transactions` are created automatically on startup. Amounts are stored as integer cents.

Inspect the DB: `sqlite3 data/miebank.db "SELECT id, name, email, balance_cents FROM users;"`

## Test
```bash
go test ./...
```

## API
| Method | Path | Body |
| --- | --- | --- |
| POST | `/api/signup` | `{name, email, password}` |
| POST | `/api/login` | `{email, password}` |
| POST | `/api/logout` | – |
| GET | `/api/me` | – |
| GET | `/api/balance` | – |
| POST | `/api/deposit` | `{amount}` |
| POST | `/api/withdraw` | `{amount}` |
| GET | `/api/transactions` | – |

## Structure
```
main.go        routes & server
handlers.go    HTTP handlers + validation
store.go       SQLite persistence (schema + queries)
session.go     in-memory session store
static/        index.html (login/signup), dashboard.html, about.html, css/, js/
```
