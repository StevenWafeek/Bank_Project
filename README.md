# MIeBank

A simple banking web app: **Go** backend + **React** frontend (Vite).

## Features
- Sign up / log in (bcrypt-hashed passwords, HTTP-only session cookies, 30-min sliding sessions)
- Dashboard with 3 options: **Your Balance**, **Withdraw**, **Deposit**
- Deposits and withdrawals must be whole dollar amounts (no decimals), at least $20 (max $1,000,000 per transaction)
- Recent transaction history
- About page for MIeBank

## Run
Requires Go 1.22+ and Node.js 18+.

```bash
# 1. Build the React frontend (once, and after frontend changes)
cd frontend
npm install
npm run build
cd ..

# 2. Start the server (serves the API and the built frontend)
go run .
# open http://localhost:8080
```

### Frontend development (hot reload)
Run `go run .` in one terminal and `cd frontend && npm run dev` in another, then open http://localhost:5173. Vite proxies `/api` to the Go server on :8080.

Environment variables (optional):
| Var | Default |
| --- | --- |
| `ADDR` | `:8080` |
| `DB_PATH` | `data/miebank.db` |
| `STATIC_DIR` | `frontend/dist` |

Data is stored in **SQLite** (pure-Go driver `modernc.org/sqlite`, no CGO/C compiler needed). Tables `users` and `transactions` are created automatically on startup. Amounts are stored as integer cents.

## Test
```bash
go test ./...                 # backend
cd frontend && npm test       # frontend amount validation
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
main.go                    wires config, database, API and frontend together
internal/
  config/                  environment variables
  models/                  User and Transaction types
  store/                   SQLite: schema.go, users.go, transactions.go
  session/                 in-memory login sessions
  money/                   amount rules ($20 min, whole dollars, $1M max)
  api/                     HTTP API
    server.go              route table
    middleware.go          auth guard, session cookie, request logging
    auth.go                signup / login / logout / me
    account.go             balance / deposit / withdraw / transactions
    respond.go             JSON helpers
  web/                     serves the React build (with SPA fallback)
frontend/
  src/
    api/client.js          fetch wrapper for the API
    context/AuthContext    logged-in user state
    pages/                 AuthPage, DashboardPage, AboutPage
    components/            Navbar, Footer, Layout, RequireAuth, Message
      auth/                LoginForm, SignupForm
      dashboard/           OptionCards, BalancePanel, MoneyForm, TransactionsTable
    utils/                 money formatting and amount validation
    styles/style.css
```
