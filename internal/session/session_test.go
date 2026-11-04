package session

import (
	"testing"
	"time"
)

func TestSessionExpiryAndDelete(t *testing.T) {
	m := NewManager()
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
