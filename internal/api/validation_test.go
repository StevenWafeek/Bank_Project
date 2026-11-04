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
