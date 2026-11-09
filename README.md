# MIeBank

A modern banking web app: **Go** backend + **React** frontend (Vite).

## Features
- Sign up / log in (bcrypt-hashed passwords, HTTP-only session cookies, 30-min sliding sessions)
- Interactive dashboard with quick actions:
  - **Your Balance** (with quick-access QR toggle and instant action pills)
  - **Transfer to Another Account** (instant peer transfers by account number or email, real-time recipient verification, QR code data parser/scanner)
  - **Deposit** and **Withdraw** (whole dollar amounts, $20 min, $1M max)
  - **Bill Payment** (utilities, electric, water, fiber internet, mobile phone, and health insurance)
  - **Charity Donations** (curated verified organizations, impact causes, dedicated personal notes, and official tax receipt reference)
  - **Manage Account** (profile name editing, password changes, account limits, and 1-click account copying)
- **Unique Account QR Code for Every Account** (dynamically generated for each user account, downloadable PNG, touch-free transfer sharing)
- Recent transaction history with detailed transaction types, descriptions/notes, and category filter pills
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

Data is stored in **SQLite** (pure-Go driver `modernc.org/sqlite`, no CGO/C compiler needed). Tables `users` and `transactions` are created and migrated automatically on startup. Amounts are stored as integer cents.

## Test
```bash
go test -v ./...              # backend (store, api, money, session, web)
cd frontend && npm test       # frontend validation tests
```

## API
| Method | Path | Body | Description |
| --- | --- | --- | --- |
| POST | `/api/signup` | `{name, email, password}` | Register new account |
| POST | `/api/login` | `{email, password}` | Log in |
| POST | `/api/logout` | – | Log out |
| GET | `/api/me` | – | Current authenticated user profile |
| GET | `/api/balance` | – | Current balance and account number |
| POST | `/api/deposit` | `{amount}` | Deposit funds ($20 min, whole dollars) |
| POST | `/api/withdraw` | `{amount}` | Withdraw funds ($20 min, whole dollars) |
| POST | `/api/transfer` | `{recipient, amount, note}` | Instant transfer to another MIeBank account |
| POST | `/api/transfer/lookup` | `{query}` | Verify recipient before transferring |
| POST | `/api/bills/pay` | `{biller, reference, amount}` | Pay utility or service bill |
| POST | `/api/donate` | `{charity, amount, note, anonymous}` | Donate to charity organization |
| PUT | `/api/account/profile` | `{name}` | Update user profile name |
| PUT | `/api/account/password` | `{current_password, new_password}` | Update account password |
| GET | `/api/account/qr` | – | Unique account QR code image (PNG) |
| GET | `/api/transactions` | – | Transaction history with descriptions |

## Structure
```
main.go                    wires config, database, API and frontend together
internal/
  config/                  environment variables
  models/                  User and Transaction types
  store/                   SQLite: schema.go, store.go, users.go, transactions.go
  session/                 in-memory login sessions
  money/                   amount validation rules
  api/                     HTTP API
    server.go              route table
    middleware.go          auth guard, session cookie, request logging
    auth.go                signup / login / logout / me
    account.go             balance / deposit / withdraw / transfer / bills / donate / qr
    respond.go             JSON helpers
  web/                     serves the React build (with SPA fallback)
frontend/
  src/
    api/client.js          fetch wrapper for the API
    context/AuthContext    logged-in user state
    pages/                 AuthPage, DashboardPage, AboutPage
    components/            Navbar, Footer, Layout, RequireAuth, Message
      auth/                LoginForm, SignupForm
      dashboard/           OptionCards, BalancePanel, MoneyForm, TransferPanel,
                           BillsPanel, DonatePanel, ManageAccountPanel,
                           AccountQR, TransactionsTable
    utils/                 money formatting and amount validation
    styles/style.css
```
