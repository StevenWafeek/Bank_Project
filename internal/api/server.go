// Package api implements the MIeBank JSON HTTP API.
package api

import (
	"net/http"

	"miebank/internal/session"
	"miebank/internal/store"
)

type Server struct {
	store    *store.Store
	sessions *session.Manager
}

func NewServer(st *store.Store, sessions *session.Manager) *Server {
	return &Server{store: st, sessions: sessions}
}

// Register adds all API routes to mux.
func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/signup", s.handleSignup)
	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.handleLogout)
	mux.HandleFunc("GET /api/me", s.requireAuth(s.handleMe))
	mux.HandleFunc("GET /api/balance", s.requireAuth(s.handleBalance))
	mux.HandleFunc("POST /api/deposit", s.requireAuth(s.handleDeposit))
	mux.HandleFunc("POST /api/withdraw", s.requireAuth(s.handleWithdraw))
	mux.HandleFunc("GET /api/transactions", s.requireAuth(s.handleTransactions))
}
