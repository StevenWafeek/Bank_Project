package main

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const sessionTTL = 30 * time.Minute

type session struct {
	userID    int64
	expiresAt time.Time
}

type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{sessions: map[string]session{}}
}

func (m *SessionManager) Create(userID int64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[token] = session{userID: userID, expiresAt: time.Now().Add(sessionTTL)}
	return token, nil
}

// Get returns the user for a token and slides its expiry forward.
func (m *SessionManager) Get(token string) (int64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[token]
	if !ok {
		return 0, false
	}
	if time.Now().After(s.expiresAt) {
		delete(m.sessions, token)
		return 0, false
	}
	s.expiresAt = time.Now().Add(sessionTTL)
	m.sessions[token] = s
	return s.userID, true
}

func (m *SessionManager) Delete(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
}
