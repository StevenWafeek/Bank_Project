package api

import (
	"log"
	"net/http"

	"miebank/internal/models"
	"miebank/internal/session"
)

const cookieName = "miebank_session"

type authedHandler func(http.ResponseWriter, *http.Request, models.User)

// requireAuth resolves the session cookie to a user, or responds 401.
func (s *Server) requireAuth(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "not logged in")
			return
		}
		userID, ok := s.sessions.Get(c.Value)
		if !ok {
			writeError(w, http.StatusUnauthorized, "session expired, please log in again")
			return
		}
		u, err := s.store.GetByID(userID)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "not logged in")
			return
		}
		next(w, r, u)
	}
}

func (s *Server) setSessionCookie(w http.ResponseWriter, r *http.Request, userID int64) error {
	token, err := s.sessions.Create(userID)
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
		MaxAge:   int(session.TTL.Seconds()),
	})
	return nil
}

func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
