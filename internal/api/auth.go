package api

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"miebank/internal/models"
	"miebank/internal/store"
)

func generateAccountNumber() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000_0000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("MIE%010d", n.Int64()), nil
}

func validateSignup(name, email, password string) error {
	if len(name) < 2 || len(name) > 60 {
		return errors.New("name must be between 2 and 60 characters")
	}
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		return errors.New("please enter a valid email address")
	}
	if len(password) < 8 || len(password) > 72 {
		return errors.New("password must be between 8 and 72 characters")
	}
	return nil
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := validateSignup(req.Name, req.Email, req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	acct, err := generateAccountNumber()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	u, err := s.store.CreateUser(req.Name, req.Email, string(hash), acct)
	if errors.Is(err, store.ErrUserExists) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		log.Printf("create user: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := s.setSessionCookie(w, r, u.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(u))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	u, err := s.store.GetByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err := s.setSessionCookie(w, r, u.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		s.sessions.Delete(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request, u models.User) {
	writeJSON(w, http.StatusOK, toUserResponse(u))
}
