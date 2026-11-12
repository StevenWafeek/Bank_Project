package api

import (
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
)

func TestSignupValidation(t *testing.T) {
	srv := newTestServer(t)
	c := &http.Client{}
	withSignupField := func(field, value string) map[string]string {
		body := signupRequest("Alice", "a@example.com", "password123", "5342000000000001", "EG96MIEB0000000000000001")
		body[field] = value
		return body
	}
	cases := map[string]map[string]string{
		"short name":                  withSignupField("name", "A"),
		"bad email":                   withSignupField("email", "not-an-email"),
		"email without domain suffix": withSignupField("email", "alice@gmail"),
		"short password":              withSignupField("password", "short"),
		"long password":               withSignupField("password", strings.Repeat("x", 73)),
		"missing card number":         withSignupField("card_number", ""),
		"invalid card number":         withSignupField("card_number", "1234"),
		"invalid card expiry":         withSignupField("card_expiry", "13/30"),
		"invalid IBAN":                withSignupField("iban", "not-an-iban"),
	}
	for name, body := range cases {
		if code, _ := post(t, c, srv.URL+"/api/signup", body); code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", name, code)
		}
	}
	if code, _ := post(t, c, srv.URL+"/api/signup", map[string]any{"name": "Alice", "unknown": 1}); code != http.StatusBadRequest {
		t.Errorf("unknown field: expected 400, got %d", code)
	}
}

func TestMeAndLogout(t *testing.T) {
	srv := newTestServer(t)
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar}

	if code, _ := post(t, c, srv.URL+"/api/signup", signupRequest("Bob", "bob@example.com", "password123", "5342000000000002", "EG96MIEB0000000000000002")); code != http.StatusCreated {
		t.Fatalf("signup: %d", code)
	}
	res, err := c.Get(srv.URL + "/api/me")
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("me: %v %v", err, res.StatusCode)
	}
	res.Body.Close()

	if code, _ := post(t, c, srv.URL+"/api/logout", nil); code != http.StatusOK {
		t.Fatalf("logout: %d", code)
	}
	res, _ = c.Get(srv.URL + "/api/me")
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("me after logout: expected 401, got %d", res.StatusCode)
	}
}
