package main

import (
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"
)

func TestParseAmount(t *testing.T) {
	cases := []struct {
		in      float64
		want    int64
		wantErr bool
	}{
		{100, 10000, false},
		{20, 2000, false},
		{20.01, 0, true},
		{20.5, 0, true},
		{999_999, 99_999_900, false},
		{10, 0, true},
		{0.01, 0, true},
		{19.99, 0, true},
		{1_000_000, 100_000_000, false},
		{0, 0, true},
		{-1, 0, true},
		{20.005, 0, true},
		{1_000_000.01, 0, true},
	}
	for _, c := range cases {
		got, err := parseAmount(c.in)
		if (err != nil) != c.wantErr || got != c.want {
			t.Errorf("parseAmount(%v) = %d, %v; want %d, err=%v", c.in, got, err, c.want, c.wantErr)
		}
	}
}

func TestSessionExpiryAndDelete(t *testing.T) {
	m := NewSessionManager()
	tok, err := m.Create(7)
	if err != nil {
		t.Fatal(err)
	}
	if id, ok := m.Get(tok); !ok || id != 7 {
		t.Fatalf("expected valid session for user 7, got %d %v", id, ok)
	}
	if _, ok := m.Get("bogus"); ok {
		t.Fatal("unknown token should be invalid")
	}
	m.sessions[tok] = session{userID: 7, expiresAt: time.Now().Add(-time.Second)}
	if _, ok := m.Get(tok); ok {
		t.Fatal("expired session should be invalid")
	}
	tok2, _ := m.Create(8)
	m.Delete(tok2)
	if _, ok := m.Get(tok2); ok {
		t.Fatal("deleted session should be invalid")
	}
}

func TestSignupValidation(t *testing.T) {
	srv := newTestServer(t)
	c := &http.Client{}
	cases := map[string]map[string]string{
		"short name":     {"name": "A", "email": "a@example.com", "password": "password123"},
		"bad email":      {"name": "Alice", "email": "not-an-email", "password": "password123"},
		"short password": {"name": "Alice", "email": "a@example.com", "password": "short"},
		"long password":  {"name": "Alice", "email": "a@example.com", "password": strings.Repeat("x", 73)},
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

	if code, _ := post(t, c, srv.URL+"/api/signup", map[string]string{"name": "Bob", "email": "bob@example.com", "password": "password123"}); code != http.StatusCreated {
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
