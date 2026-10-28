package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	cookieName       = "miebank_session"
	minTxAmountCents = 20_00        // $20.00 minimum per transaction
	maxTxAmountCents = 1_000_000_00 // $1,000,000.00 maximum per transaction
)

type App struct {
	store    *Store
	sessions *SessionManager
}

type userResponse struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	AccountNumber string `json:"account_number"`
	BalanceCents  int64  `json:"balance_cents"`
}

func toUserResponse(u User) userResponse {
	return userResponse{u.ID, u.Name, u.Email, u.AccountNumber, u.BalanceCents}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func (a *App) setSessionCookie(w http.ResponseWriter, r *http.Request, userID int64) error {
	token, err := a.sessions.Create(userID)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	return nil
}

func (a *App) requireAuth(next func(http.ResponseWriter, *http.Request, User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "not logged in")
			return
		}
		userID, ok := a.sessions.Get(c.Value)
		if !ok {
			writeError(w, http.StatusUnauthorized, "session expired, please log in again")
			return
		}
		u, err := a.store.GetByID(userID)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "not logged in")
			return
		}
		next(w, r, u)
	}
}

func generateAccountNumber() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000_0000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("MIE%010d", n.Int64()), nil
}

func (a *App) handleSignup(w http.ResponseWriter, r *http.Request) {
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

	if len(req.Name) < 2 || len(req.Name) > 60 {
		writeError(w, http.StatusBadRequest, "name must be between 2 and 60 characters")
		return
	}
	if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email {
		writeError(w, http.StatusBadRequest, "please enter a valid email address")
		return
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		writeError(w, http.StatusBadRequest, "password must be between 8 and 72 characters")
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
	u, err := a.store.CreateUser(req.Name, req.Email, string(hash), acct)
	if errors.Is(err, ErrUserExists) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		log.Printf("create user: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := a.setSessionCookie(w, r, u.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(u))
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	u, err := a.store.GetByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err := a.setSessionCookie(w, r, u.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		a.sessions.Delete(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request, u User) {
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

func (a *App) handleBalance(w http.ResponseWriter, r *http.Request, u User) {
	writeJSON(w, http.StatusOK, map[string]any{
		"balance_cents":  u.BalanceCents,
		"account_number": u.AccountNumber,
	})
}

// parseAmount converts a whole-dollar amount to cents, rejecting decimals and out-of-range values.
func parseAmount(amount float64) (int64, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return 0, errors.New("amount must be greater than 0")
	}
	if amount != math.Trunc(amount) {
		return 0, errors.New("amount must be a whole number (no decimals)")
	}
	cents := amount * 100
	if cents < minTxAmountCents {
		return 0, errors.New("the minimum amount for a deposit or withdrawal is $20.00")
	}
	if cents > maxTxAmountCents {
		return 0, errors.New("amount exceeds the per-transaction limit of 1,000,000.00")
	}
	return int64(cents), nil
}

func (a *App) handleMoney(w http.ResponseWriter, r *http.Request, u User, txType string, sign int64) {
	var req struct {
		Amount float64 `json:"amount"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	cents, err := parseAmount(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tx, err := a.store.ApplyTransaction(u.ID, txType, sign*cents)
	if errors.Is(err, ErrInsufficientFunds) {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err != nil {
		log.Printf("apply transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"transaction":   tx,
		"balance_cents": tx.BalanceCents,
	})
}

func (a *App) handleDeposit(w http.ResponseWriter, r *http.Request, u User) {
	a.handleMoney(w, r, u, "deposit", 1)
}

func (a *App) handleWithdraw(w http.ResponseWriter, r *http.Request, u User) {
	a.handleMoney(w, r, u, "withdraw", -1)
}

func (a *App) handleTransactions(w http.ResponseWriter, r *http.Request, u User) {
	txs, err := a.store.ListTransactions(u.ID, 50)
	if err != nil {
		log.Printf("list transactions: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": txs})
}
