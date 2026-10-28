package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func newTestServer(t *testing.T) *httptest.Server {
	store, err := NewStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	app := &App{store: store, sessions: NewSessionManager()}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/signup", app.handleSignup)
	mux.HandleFunc("POST /api/login", app.handleLogin)
	mux.HandleFunc("POST /api/logout", app.handleLogout)
	mux.HandleFunc("GET /api/me", app.requireAuth(app.handleMe))
	mux.HandleFunc("GET /api/balance", app.requireAuth(app.handleBalance))
	mux.HandleFunc("POST /api/deposit", app.requireAuth(app.handleDeposit))
	mux.HandleFunc("POST /api/withdraw", app.requireAuth(app.handleWithdraw))
	mux.HandleFunc("GET /api/transactions", app.requireAuth(app.handleTransactions))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func post(t *testing.T, c *http.Client, url string, body any) (int, map[string]any) {
	b, _ := json.Marshal(body)
	res, err := c.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res.StatusCode, out
}

func TestBankingFlow(t *testing.T) {
	srv := newTestServer(t)
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar}

	if res, _ := c.Get(srv.URL + "/api/balance"); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 before login, got %d", res.StatusCode)
	}

	code, _ := post(t, c, srv.URL+"/api/signup", map[string]string{"name": "Steven", "email": "s@example.com", "password": "password123"})
	if code != http.StatusCreated {
		t.Fatalf("signup: %d", code)
	}
	if code, _ := post(t, c, srv.URL+"/api/signup", map[string]string{"name": "Steven", "email": "S@example.com", "password": "password123"}); code != http.StatusConflict {
		t.Fatalf("duplicate signup: %d", code)
	}

	code, out := post(t, c, srv.URL+"/api/deposit", map[string]any{"amount": 100})
	if code != http.StatusOK || out["balance_cents"].(float64) != 10000 {
		t.Fatalf("deposit: %d %v", code, out)
	}
	code, out = post(t, c, srv.URL+"/api/withdraw", map[string]any{"amount": 40})
	if code != http.StatusOK || out["balance_cents"].(float64) != 6000 {
		t.Fatalf("withdraw: %d %v", code, out)
	}
	if code, _ := post(t, c, srv.URL+"/api/withdraw", map[string]any{"amount": 1000}); code != http.StatusUnprocessableEntity {
		t.Fatalf("overdraw: %d", code)
	}
	for _, bad := range []any{-5, 0, 1.234, 19.99, 5, 20.5, 100.01} {
		if code, _ := post(t, c, srv.URL+"/api/deposit", map[string]any{"amount": bad}); code != http.StatusBadRequest {
			t.Fatalf("bad amount %v: %d", bad, code)
		}
	}

	res, err := c.Get(srv.URL + "/api/transactions")
	if err != nil {
		t.Fatal(err)
	}
	var txOut struct {
		Transactions []Transaction `json:"transactions"`
	}
	_ = json.NewDecoder(res.Body).Decode(&txOut)
	res.Body.Close()
	if len(txOut.Transactions) != 2 || txOut.Transactions[0].Type != "withdraw" || txOut.Transactions[0].BalanceCents != 6000 {
		t.Fatalf("transactions: %+v", txOut.Transactions)
	}

	c2 := &http.Client{}
	if code, _ := post(t, c2, srv.URL+"/api/login", map[string]string{"email": "s@example.com", "password": "wrong"}); code != http.StatusUnauthorized {
		t.Fatalf("bad login: %d", code)
	}
	if code, _ := post(t, c2, srv.URL+"/api/login", map[string]string{"email": "s@example.com", "password": "password123"}); code != http.StatusOK {
		t.Fatalf("login: %d", code)
	}
}
