# MIeBank

A modern banking web app: **Go** backend + **React** frontend (Vite).

## Features
- Sign up / log in (bcrypt-hashed passwords, HTTP-only session cookies, 30-min sliding sessions; sign in with **Email** or **16-digit Card Number**)
- **Virtual Debit Card & International IBAN**:
  - Card number, expiry date (`MM/YY`), and IBAN provided during signup
  - Displayed prominently upon signing in (Balance panel) with copy controls and unmask toggles
  - Editable from the **Account Profile / Manage Account** page
- Interactive dashboard with quick actions:
  - **Your Balance** (with integrated virtual debit card, quick-access QR toggle, and instant action pills)
  - **Transfer to Another Account** (instant peer transfers by account number, email, card number, or IBAN; real-time recipient verification; QR code data parser/scanner)
  - **Deposit** and **Withdraw** (whole dollar amounts, $20 min, $1M max)
  - **Bill Payment** (utilities, electric, water, fiber internet, mobile phone, and health insurance)
  - **Charity Donations** (curated verified organizations, impact causes, dedicated personal notes, and official tax receipt reference)
  - **Manage Account** (profile name editing, password changes, card number/date/IBAN inspection, account limits, and 1-click copying)
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
| `PORT` | `8080` |
| `ADDR` | `:${PORT}` |
| `DB_PATH` | `data/miebank.db` |
| `STATIC_DIR` | `frontend/dist` |

Data is stored in **SQLite** (pure-Go driver `modernc.org/sqlite`, no CGO/C compiler needed). Tables `users` and `transactions` are created and migrated automatically on startup. Amounts are stored as integer cents.

## Deploy for free with Cloudflare

The Cloudflare Worker serves the built React app and implements the Go API against a Cloudflare D1 database. The Worker version uses PBKDF2 password hashes and stores sessions in D1; it is a separate backend implementation, and existing SQLite accounts are not migrated.

Use Cloudflare's Free plan only. If Cloudflare asks for payment details or a paid upgrade, stop rather than continuing. This demo is not suitable for real banking, card, or personal data.

1. Install [Node.js 22.12 or later](https://nodejs.org/en/download) and create a free Cloudflare account. Wrangler 4 requires this Node.js version.
2. From the repository root, install the Worker and frontend dependencies:

   ```powershell
   npm.cmd install
   npm.cmd --prefix frontend ci
   ```
3. Log in and create a free D1 database:

   ```powershell
   npx.cmd wrangler login
   npx.cmd wrangler d1 create miebank-db
   ```

   Copy the `database_id` printed by the command into `wrangler.toml`, replacing the all-zero local-development placeholder.
4. Apply the schema to the remote database and deploy:

   ```powershell
   npx.cmd wrangler d1 migrations apply miebank-db --remote
   npm.cmd run cf:deploy
   ```

5. Open the `workers.dev` URL printed by Wrangler and use fictional test accounts only.

For local development, build the frontend, create a local D1 database, then start Wrangler:

```powershell
npx.cmd wrangler d1 migrations apply miebank-db --local
npm.cmd run cf:dev
```

The frontend and `/api` endpoints share the same origin. The local and remote D1 databases are separate, so test data created locally is not deployed.

## Deploy with Docker on Render

Create a Render Web Service from this repository, select **Docker**, and use the repository root as the Dockerfile path (`./Dockerfile`). The image builds the frontend and Go server together; the server listens on Render's `PORT`.

Render may require card verification before creating a service. Do not continue if you don't want to provide payment details. SQLite data is not persistent on an ephemeral filesystem. Attach a persistent disk to the service, mount it at `/var/data`, and set `DB_PATH` to `/var/data/miebank.db` in the service's environment variables. Without a persistent disk, database contents can be lost when the service restarts or is redeployed.

## Deploy with Firebase Hosting and Cloud Run

Firebase Hosting serves the built React app, while its `/api/**` rewrite forwards API requests to the Go service on Cloud Run. The rewrite in `firebase.json` expects a Cloud Run service named `miebank-api` in `us-central1`; use that same region when deploying the service, or update the rewrite to match your chosen region.

The Firebase Console's JavaScript SDK snippet is for using Firebase products such as Firebase Authentication or Analytics in the frontend. This app currently uses its Go API for authentication and data, so that snippet is not needed to deploy it.

Prerequisites: Node.js/npm, Google Cloud CLI, and a Firebase project with billing enabled. The Firebase CLI can be run through `npx` below.
In Windows PowerShell, use the `.cmd` launchers shown below to avoid execution-policy blocks on unsigned `.ps1` wrappers.

```powershell
# Authenticate and select the same Google Cloud/Firebase project
npx.cmd firebase-tools login
gcloud auth login
gcloud config set project miebankapp
# Firebase Hosting uses the default project configured in .firebaserc

# Build and deploy the Go app from the repository root
gcloud run deploy miebank-api --source . --region us-central1 --allow-unauthenticated

# Build and deploy the frontend to Firebase Hosting
npm.cmd --prefix frontend ci
npm.cmd --prefix frontend run build
npx.cmd firebase-tools deploy --only hosting
```

The Cloud Run service is publicly reachable so Firebase Hosting can forward requests to it; application endpoints still use the app's own login. The current SQLite database and in-memory sessions are not suitable for persistent, multi-instance production use on Cloud Run. Do not store real account data until the app uses persistent database storage and shared session storage.

## Test
```bash
go test -v ./...              # backend (store, api, money, session, web)
cd frontend && npm test       # frontend validation tests
```

## API
| Method | Path | Body | Description |
| --- | --- | --- | --- |
| POST | `/api/signup` | `{name, email, password, card_number, card_expiry, iban}` | Register new account |
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
| PUT | `/api/account/card` | `{card_number, card_expiry, iban}` | Update card and banking details |
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
