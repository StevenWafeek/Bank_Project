// Package session keeps logged-in sessions in memory, keyed by random tokens.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// TTL is how long a session stays valid without activity.
const TTL = 30 * time.Minute

type session struct {
	userID    int64
	expiresAt time.Time
}

type Manager struct {
	mu       sync.Mutex
	sessions map[string]session
}

func NewManager() *Manager {
	return &Manager{sessions: map[string]session{}}
}

func (m *Manager) Create(userID int64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[token] = session{userID: userID, expiresAt: time.Now().Add(TTL)}
	return token, nil
}

// Get returns the user for a token and slides its expiry forward.
func (m *Manager) Get(token string) (int64, bool) {
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
	s.expiresAt = time.Now().Add(TTL)
	m.sessions[token] = s
	return s.userID, true
}

func (m *Manager) Delete(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
}
