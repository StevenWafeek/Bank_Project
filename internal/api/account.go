package api

import (
	"errors"
	"log"
	"net/http"

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
