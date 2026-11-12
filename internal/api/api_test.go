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

func signupRequest(name, email, password, cardNumber, iban string) map[string]string {
	return map[string]string{
		"name":        name,
		"email":       email,
		"password":    password,
		"card_number": cardNumber,
		"card_expiry": "12/30",
		"iban":        iban,
	}
}

func TestBankingFlow(t *testing.T) {
	srv := newTestServer(t)
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar}

	if res, _ := c.Get(srv.URL + "/api/balance"); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 before login, got %d", res.StatusCode)
	}

	code, _ := post(t, c, srv.URL+"/api/signup", signupRequest("Steven", "s@example.com", "password123", "5342000000000001", "EG96MIEB0000000000000001"))
	if code != http.StatusCreated {
		t.Fatalf("signup: %d", code)
	}
	if code, _ := post(t, c, srv.URL+"/api/signup", signupRequest("Steven", "S@example.com", "password123", "5342000000000001", "EG96MIEB0000000000000001")); code != http.StatusConflict {
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
	code, u1 := post(t, c1, srv.URL+"/api/signup", signupRequest("Alice", "alice@bank.com", "password123", "5342000000000002", "EG96MIEB0000000000000002"))
	if code != http.StatusCreated {
		t.Fatalf("signup user1: %d", code)
	}
	acct1 := u1["account_number"].(string)

	// User 2 (Bob)
	jar2, _ := cookiejar.New(nil)
	c2 := &http.Client{Jar: jar2}
	code, u2 := post(t, c2, srv.URL+"/api/signup", signupRequest("Bob", "bob@bank.com", "password123", "5342000000000003", "EG96MIEB0000000000000003"))
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

func TestCardAndIBANFeatures(t *testing.T) {
	srv := newTestServer(t)

	// User 1
	jar1, _ := cookiejar.New(nil)
	c1 := &http.Client{Jar: jar1}
	code, u1 := post(t, c1, srv.URL+"/api/signup", signupRequest("Evelyn Reed", "evelyn@bank.com", "password123", "5342000000000004", "EG96MIEB0000000000000004"))
	if code != http.StatusCreated {
		t.Fatalf("signup user1: %d %v", code, u1)
	}

	card1, ok1 := u1["card_number"].(string)
	exp1, ok2 := u1["card_expiry"].(string)
	iban1, ok3 := u1["iban"].(string)

	if !ok1 || len(card1) != 16 {
		t.Fatalf("expected 16-digit card number, got %v", u1["card_number"])
	}
	if !ok2 || len(exp1) != 5 {
		t.Fatalf("expected MM/YY card expiry, got %v", u1["card_expiry"])
	}
	if !ok3 || len(iban1) < 20 {
		t.Fatalf("expected valid IBAN, got %v", u1["iban"])
	}

	// User 2
	jar2, _ := cookiejar.New(nil)
	c2 := &http.Client{Jar: jar2}
	code, u2 := post(t, c2, srv.URL+"/api/signup", signupRequest("Marcus Vance", "marcus@bank.com", "password123", "5342000000000005", "EG96MIEB0000000000000005"))
	if code != http.StatusCreated {
		t.Fatalf("signup user2: %d %v", code, u2)
	}
	card2 := u2["card_number"].(string)
	iban2 := u2["iban"].(string)

	// Test login with card number (clean 16 digits)
	c3 := &http.Client{}
	code, loginOut := post(t, c3, srv.URL+"/api/login", map[string]string{
		"email":    card1,
		"password": "password123",
	})
	if code != http.StatusOK || loginOut["name"] != "Evelyn Reed" {
		t.Fatalf("login with card_number: %d %v", code, loginOut)
	}

	// Test login with card number containing spaces (e.g. 5342 0000 ...)
	formattedCard := card1[0:4] + " " + card1[4:8] + " " + card1[8:12] + " " + card1[12:16]
	code, loginOut2 := post(t, c3, srv.URL+"/api/login", map[string]string{
		"email":    formattedCard,
		"password": "password123",
	})
	if code != http.StatusOK || loginOut2["name"] != "Evelyn Reed" {
		t.Fatalf("login with spaced card_number: %d %v", code, loginOut2)
	}

	// Test login with IBAN
	code, loginOut3 := post(t, c3, srv.URL+"/api/login", map[string]string{
		"email":    iban1,
		"password": "password123",
	})
	if code != http.StatusOK || loginOut3["name"] != "Evelyn Reed" {
		t.Fatalf("login with iban: %d %v", code, loginOut3)
	}

	// Test /api/me returns card details
	res, err := c1.Get(srv.URL + "/api/me")
	if err != nil {
		t.Fatal(err)
	}
	var meOut map[string]any
	_ = json.NewDecoder(res.Body).Decode(&meOut)
	res.Body.Close()
	if meOut["card_number"] != card1 || meOut["card_expiry"] != exp1 || meOut["iban"] != iban1 {
		t.Fatalf("api/me card details mismatch: %v", meOut)
	}

	// Test /api/balance returns card details
	resBal, err := c1.Get(srv.URL + "/api/balance")
	if err != nil {
		t.Fatal(err)
	}
	var balOut map[string]any
	_ = json.NewDecoder(resBal.Body).Decode(&balOut)
	resBal.Body.Close()
	if balOut["card_number"] != card1 || balOut["card_expiry"] != exp1 || balOut["iban"] != iban1 {
		t.Fatalf("api/balance card details mismatch: %v", balOut)
	}

	// Fund Evelyn with $300
	post(t, c1, srv.URL+"/api/deposit", map[string]any{"amount": 300})

	// Lookup Marcus using his card number
	code, outLookupCard := post(t, c1, srv.URL+"/api/transfer/lookup", map[string]string{"query": card2})
	if code != http.StatusOK || outLookupCard["name"] != "Marcus Vance" {
		t.Fatalf("lookup by card: %d %v", code, outLookupCard)
	}

	// Lookup Marcus using his IBAN
	code, outLookupIBAN := post(t, c1, srv.URL+"/api/transfer/lookup", map[string]string{"query": iban2})
	if code != http.StatusOK || outLookupIBAN["name"] != "Marcus Vance" {
		t.Fatalf("lookup by iban: %d %v", code, outLookupIBAN)
	}

	// Transfer $45 to Marcus using his card number
	code, txCard := post(t, c1, srv.URL+"/api/transfer", map[string]any{
		"recipient": card2,
		"amount":    45,
		"note":      "Transfer via card",
	})
	if code != http.StatusOK || txCard["balance_cents"].(float64) != 25500 {
		t.Fatalf("transfer via card: %d %v", code, txCard)
	}

	// Transfer $35 to Marcus using his IBAN
	code, txIBAN := post(t, c1, srv.URL+"/api/transfer", map[string]any{
		"recipient": iban2,
		"amount":    35,
		"note":      "Transfer via IBAN",
	})
	if code != http.StatusOK || txIBAN["balance_cents"].(float64) != 22000 {
		t.Fatalf("transfer via iban: %d %v", code, txIBAN)
	}

	code, out := put(t, c1, srv.URL+"/api/account/card", map[string]string{
		"card_number": "5342000000009999",
		"card_expiry": "11/31",
		"iban":        "EG96MIEB0000000099999999",
	})
	if code != http.StatusOK || out["card_number"] != "5342000000009999" ||
		out["card_expiry"] != "11/31" || out["iban"] != "EG96MIEB0000000099999999" {
		t.Fatalf("update card details: %d %v", code, out)
	}

	code, out = put(t, c1, srv.URL+"/api/account/card", map[string]string{
		"card_number": card2,
		"card_expiry": "11/31",
		"iban":        "EG96MIEB0000000099999999",
	})
	if code != http.StatusConflict || out["error"] == nil {
		t.Fatalf("duplicate card number: %d %v", code, out)
	}

	res, err = c1.Get(srv.URL + "/api/me")
	if err != nil {
		t.Fatal(err)
	}
	var updatedMe map[string]any
	if err := json.NewDecoder(res.Body).Decode(&updatedMe); err != nil {
		res.Body.Close()
		t.Fatal(err)
	}
	res.Body.Close()
	if updatedMe["card_number"] != "5342000000009999" ||
		updatedMe["card_expiry"] != "11/31" || updatedMe["iban"] != "EG96MIEB0000000099999999" {
		t.Fatalf("updated profile card details mismatch: %v", updatedMe)
	}
}
