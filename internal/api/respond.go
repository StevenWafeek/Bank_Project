package api

import (
	"encoding/json"
	"net/http"
	"time"

	"miebank/internal/models"
)

type userResponse struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	AccountNumber string    `json:"account_number"`
	BalanceCents  int64     `json:"balance_cents"`
	CreatedAt     time.Time `json:"created_at"`
}

func toUserResponse(u models.User) userResponse {
	return userResponse{u.ID, u.Name, u.Email, u.AccountNumber, u.BalanceCents, u.CreatedAt}
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
