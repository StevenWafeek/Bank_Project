package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"miebank/internal/models"
	"miebank/internal/session"
	"miebank/internal/store"
)

func newTestServer(t *testing.T) *httptest.Server {
	st, err := store.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	mux := http.NewServeMux()
	NewServer(st, session.NewManager()).Register(mux)
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

func put(t *testing.T, c *http.Client, url string, body any) (int, map[string]any) {
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)
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
		Transactions []models.Transaction `json:"transactions"`
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

func TestTransferAndAccountFeatures(t *testing.T) {
	srv := newTestServer(t)

	// User 1 (Alice)
	jar1, _ := cookiejar.New(nil)
	c1 := &http.Client{Jar: jar1}
	code, u1 := post(t, c1, srv.URL+"/api/signup", map[string]string{"name": "Alice", "email": "alice@bank.com", "password": "password123"})
	if code != http.StatusCreated {
		t.Fatalf("signup user1: %d", code)
	}
	acct1 := u1["account_number"].(string)

	// User 2 (Bob)
	jar2, _ := cookiejar.New(nil)
	c2 := &http.Client{Jar: jar2}
	code, u2 := post(t, c2, srv.URL+"/api/signup", map[string]string{"name": "Bob", "email": "bob@bank.com", "password": "password123"})
	if code != http.StatusCreated {
		t.Fatalf("signup user2: %d", code)
	}
	acct2 := u2["account_number"].(string)

	// Deposit $200 for Alice
	post(t, c1, srv.URL+"/api/deposit", map[string]any{"amount": 200})

	// Lookup Bob from Alice
	code, out := post(t, c1, srv.URL+"/api/transfer/lookup", map[string]string{"query": acct2})
	if code != http.StatusOK || out["name"] != "Bob" {
		t.Fatalf("lookup: %d %v", code, out)
	}

	// Lookup self should fail
	code, _ = post(t, c1, srv.URL+"/api/transfer/lookup", map[string]string{"query": acct1})
	if code != http.StatusBadRequest {
		t.Fatalf("self lookup should fail: %d", code)
	}

	// Transfer $50 to Bob
	code, out = post(t, c1, srv.URL+"/api/transfer", map[string]any{"recipient": acct2, "amount": 50, "note": "Lunch"})
	if code != http.StatusOK || out["balance_cents"].(float64) != 15000 {
		t.Fatalf("transfer: %d %v", code, out)
	}

	// Verify Bob received $50
	res, _ := c2.Get(srv.URL + "/api/balance")
	var bal2 map[string]any
	_ = json.NewDecoder(res.Body).Decode(&bal2)
	res.Body.Close()
	if bal2["balance_cents"].(float64) != 5000 {
		t.Fatalf("bob balance: %v", bal2)
	}

	// Pay a bill ($30)
	code, out = post(t, c1, srv.URL+"/api/bills/pay", map[string]any{"biller": "Metro Power", "reference": "BILL-1002", "amount": 30})
	if code != http.StatusOK || out["balance_cents"].(float64) != 12000 || out["confirmation_code"] == "" {
		t.Fatalf("pay bill: %d %v", code, out)
	}

	// Make a donation ($20)
	code, out = post(t, c1, srv.URL+"/api/donate", map[string]any{"charity": "Clean Water Initiative", "amount": 20, "note": "All the best"})
	if code != http.StatusOK || out["balance_cents"].(float64) != 10000 || out["receipt_id"] == "" {
		t.Fatalf("donate: %d %v", code, out)
	}

	// Update profile
	code, out = put(t, c1, srv.URL+"/api/account/profile", map[string]string{"name": "Alice Wonderland"})
	if code != http.StatusOK || out["name"] != "Alice Wonderland" {
		t.Fatalf("update profile: %d %v", code, out)
	}

	// Update password
	code, _ = put(t, c1, srv.URL+"/api/account/password", map[string]string{"current_password": "wrong", "new_password": "newpassword123"})
	if code != http.StatusUnauthorized {
		t.Fatalf("update password with wrong old pass: %d", code)
	}
	code, _ = put(t, c1, srv.URL+"/api/account/password", map[string]string{"current_password": "password123", "new_password": "newpassword123"})
	if code != http.StatusOK {
		t.Fatalf("update password success: %d", code)
	}

	// Check QR code endpoint
	resQR, err := c1.Get(srv.URL + "/api/account/qr")
	if err != nil || resQR.StatusCode != http.StatusOK || resQR.Header.Get("Content-Type") != "image/png" {
		t.Fatalf("qr endpoint: %v %d %s", err, resQR.StatusCode, resQR.Header.Get("Content-Type"))
	}
	resQR.Body.Close()
}
