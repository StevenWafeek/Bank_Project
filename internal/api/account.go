package api

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"

	"miebank/internal/models"
	"miebank/internal/money"
	"miebank/internal/store"
)

const historyLimit = 50

func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request, u models.User) {
	writeJSON(w, http.StatusOK, map[string]any{
		"balance_cents":  u.BalanceCents,
		"account_number": u.AccountNumber,
	})
}

func (s *Server) handleDeposit(w http.ResponseWriter, r *http.Request, u models.User) {
	s.handleMoney(w, r, u, models.TxDeposit, 1)
}

func (s *Server) handleWithdraw(w http.ResponseWriter, r *http.Request, u models.User) {
	s.handleMoney(w, r, u, models.TxWithdraw, -1)
}

func (s *Server) handleMoney(w http.ResponseWriter, r *http.Request, u models.User, txType string, sign int64) {
	var req struct {
		Amount float64 `json:"amount"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	cents, err := money.ParseAmount(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tx, err := s.store.ApplyTransaction(u.ID, txType, sign*cents)
	if errors.Is(err, store.ErrInsufficientFunds) {
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

func (s *Server) handleTransactions(w http.ResponseWriter, r *http.Request, u models.User) {
	txs, err := s.store.ListTransactions(u.ID, historyLimit)
	if err != nil {
		log.Printf("list transactions: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": txs})
}

// handleTransfer handles sending money from the authenticated user to another account.
func (s *Server) handleTransfer(w http.ResponseWriter, r *http.Request, u models.User) {
	var req struct {
		Recipient string  `json:"recipient"`
		Amount    float64 `json:"amount"`
		Note      string  `json:"note"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Recipient = strings.TrimSpace(req.Recipient)
	if req.Recipient == "" {
		writeError(w, http.StatusBadRequest, "recipient account number or email is required")
		return
	}
	cents, err := money.ParsePaymentAmount(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tx, recipient, err := s.store.Transfer(u.ID, req.Recipient, cents, req.Note)
	if errors.Is(err, store.ErrRecipientNotFound) {
		writeError(w, http.StatusNotFound, "recipient account not found")
		return
	}
	if errors.Is(err, store.ErrSelfTransfer) {
		writeError(w, http.StatusBadRequest, "cannot transfer to your own account")
		return
	}
	if errors.Is(err, store.ErrInsufficientFunds) {
		writeError(w, http.StatusUnprocessableEntity, "insufficient funds for transfer")
		return
	}
	if err != nil {
		log.Printf("transfer: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"transaction":       tx,
		"balance_cents":     tx.BalanceCents,
		"recipient_name":    recipient.Name,
		"recipient_account": recipient.AccountNumber,
	})
}

// handleTransferLookup verifies a recipient account number or email before sending.
func (s *Server) handleTransferLookup(w http.ResponseWriter, r *http.Request, u models.User) {
	var req struct {
		Query string `json:"query"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "please enter an account number or email")
		return
	}

	recipient, err := s.store.GetByEmailOrAccount(req.Query)
	if errors.Is(err, store.ErrUserNotFound) {
		writeError(w, http.StatusNotFound, "recipient not found")
		return
	}
	if err != nil {
		log.Printf("recipient lookup: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if recipient.ID == u.ID {
		writeError(w, http.StatusBadRequest, "cannot transfer to your own account")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"name":           recipient.Name,
		"account_number": recipient.AccountNumber,
	})
}

// handlePayBill processes bill payments.
func (s *Server) handlePayBill(w http.ResponseWriter, r *http.Request, u models.User) {
	var req struct {
		Biller    string  `json:"biller"`
		Reference string  `json:"reference"`
		Amount    float64 `json:"amount"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Biller = strings.TrimSpace(req.Biller)
	req.Reference = strings.TrimSpace(req.Reference)
	if len(req.Biller) < 2 || len(req.Biller) > 80 {
		writeError(w, http.StatusBadRequest, "please select or enter a valid biller name")
		return
	}
	if len(req.Reference) < 2 || len(req.Reference) > 60 {
		writeError(w, http.StatusBadRequest, "please enter a valid account or reference number")
		return
	}
	cents, err := money.ParsePaymentAmount(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tx, err := s.store.PayBill(u.ID, req.Biller, req.Reference, cents)
	if errors.Is(err, store.ErrInsufficientFunds) {
		writeError(w, http.StatusUnprocessableEntity, "insufficient funds to pay bill")
		return
	}
	if err != nil {
		log.Printf("pay bill: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	randNum, _ := rand.Int(rand.Reader, big.NewInt(90000000))
	confirmationCode := fmt.Sprintf("BILL-%08d", randNum.Int64()+10000000)

	writeJSON(w, http.StatusOK, map[string]any{
		"transaction":       tx,
		"balance_cents":     tx.BalanceCents,
		"confirmation_code": confirmationCode,
		"biller":            req.Biller,
		"reference":         req.Reference,
	})
}

// handleDonate processes charitable donations.
func (s *Server) handleDonate(w http.ResponseWriter, r *http.Request, u models.User) {
	var req struct {
		Charity   string  `json:"charity"`
		Amount    float64 `json:"amount"`
		Note      string  `json:"note"`
		Anonymous bool    `json:"anonymous"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Charity = strings.TrimSpace(req.Charity)
	if len(req.Charity) < 2 || len(req.Charity) > 80 {
		writeError(w, http.StatusBadRequest, "please select or enter a valid charity name")
		return
	}
	cents, err := money.ParsePaymentAmount(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tx, err := s.store.Donate(u.ID, req.Charity, req.Note, cents)
	if errors.Is(err, store.ErrInsufficientFunds) {
		writeError(w, http.StatusUnprocessableEntity, "insufficient funds for donation")
		return
	}
	if err != nil {
		log.Printf("donation: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	randNum, _ := rand.Int(rand.Reader, big.NewInt(90000000))
	receiptID := fmt.Sprintf("DON-%08d", randNum.Int64()+10000000)

	writeJSON(w, http.StatusOK, map[string]any{
		"transaction":   tx,
		"balance_cents": tx.BalanceCents,
		"receipt_id":    receiptID,
		"charity":       req.Charity,
	})
}

// handleUpdateProfile allows updating user's full name.
func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request, u models.User) {
	var req struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 2 || len(req.Name) > 60 {
		writeError(w, http.StatusBadRequest, "name must be between 2 and 60 characters")
		return
	}

	if err := s.store.UpdateUserName(u.ID, req.Name); err != nil {
		log.Printf("update profile: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	u.Name = req.Name
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

// handleUpdatePassword allows the authenticated user to update their password.
func (s *Server) handleUpdatePassword(w http.ResponseWriter, r *http.Request, u models.User) {
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.CurrentPassword)) != nil {
		writeError(w, http.StatusUnauthorized, "incorrect current password")
		return
	}
	if len(req.NewPassword) < 8 || len(req.NewPassword) > 72 {
		writeError(w, http.StatusBadRequest, "new password must be between 8 and 72 characters")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := s.store.UpdateUserPassword(u.ID, string(hash)); err != nil {
		log.Printf("update password: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password updated successfully"})
}

// handleAccountQR returns a PNG QR code encoding the user's account details.
func (s *Server) handleAccountQR(w http.ResponseWriter, r *http.Request, u models.User) {
	payloadMap := map[string]string{
		"bank":           "MIeBank",
		"account_number": u.AccountNumber,
		"name":           u.Name,
	}
	payloadBytes, _ := json.Marshal(payloadMap)

	pngBytes, err := qrcode.Encode(string(payloadBytes), qrcode.Medium, 256)
	if err != nil {
		log.Printf("generate qr: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to generate QR code")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, max-age=60")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pngBytes)
}
