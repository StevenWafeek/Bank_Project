package api

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

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
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return errors.New("please enter a valid email address")
	}
	domain := email[strings.LastIndex(email, "@")+1:]
	if dot := strings.LastIndex(domain, "."); dot <= 0 || dot == len(domain)-1 {
		return errors.New("please enter a valid email address")
	}
	if len(password) < 8 || len(password) > 72 {
		return errors.New("password must be between 8 and 72 characters")
	}
	return nil
}

// reCardExpiry matches MM/YY.
var reCardExpiry = regexp.MustCompile(`^(0[1-9]|1[0-2])/\d{2}$`)

// reCardNumber matches exactly 16 digits (after stripping spaces/hyphens).
var reCardNumber = regexp.MustCompile(`^\d{16}$`)

// reIBAN matches a basic IBAN: 2 uppercase letters + 2 digits + 1–30 alphanumeric chars.
var reIBAN = regexp.MustCompile(`^[A-Z]{2}\d{2}[A-Z0-9]{1,30}$`)

func validateCardDetails(cardNumber, cardExpiry, iban string) error {
	cleanCard := store.CleanCardNumber(cardNumber)
	if !reCardNumber.MatchString(cleanCard) {
		return errors.New("card number must be exactly 16 digits")
	}
	if !reCardExpiry.MatchString(cardExpiry) {
		return errors.New("card expiry must be in MM/YY format (e.g. 08/28)")
	}
	// Expiry must not be in the past
	now := time.Now()
	parts := strings.Split(cardExpiry, "/")
	var m, y int
	fmt.Sscanf(parts[0], "%d", &m)
	fmt.Sscanf(parts[1], "%d", &y)
	y += 2000
	expTime := time.Date(y, time.Month(m+1), 1, 0, 0, 0, 0, time.UTC)
	if !expTime.After(now) {
		return errors.New("card expiry date has already passed")
	}
	cleanIBAN := store.CleanIBAN(iban)
	if !reIBAN.MatchString(cleanIBAN) {
		return errors.New("IBAN must start with 2 letters and 2 digits followed by up to 30 alphanumeric characters (e.g. EG96MIEB0000000012345678)")
	}
	return nil
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string `json:"name"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		CardNumber string `json:"card_number"`
		CardExpiry string `json:"card_expiry"`
		IBAN       string `json:"iban"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.CardExpiry = strings.TrimSpace(req.CardExpiry)

	if err := validateSignup(req.Name, req.Email, req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateCardDetails(req.CardNumber, req.CardExpiry, req.IBAN); err != nil {
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

	cleanCard := store.CleanCardNumber(req.CardNumber)
	cleanIBAN := store.CleanIBAN(req.IBAN)

	u, err := s.store.CreateUser(req.Name, req.Email, string(hash), acct, cleanCard, req.CardExpiry, cleanIBAN)
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
		Email      string `json:"email"`
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	loginID := req.Email
	if loginID == "" {
		loginID = req.Identifier
	}
	loginID = strings.TrimSpace(loginID)
	if loginID == "" {
		writeError(w, http.StatusBadRequest, "please enter your email or card number")
		return
	}
	u, err := s.store.GetByLoginIdentifier(loginID)
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
