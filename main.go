package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	addr := getenv("ADDR", ":8080")
	dbPath := getenv("DB_PATH", "data/miebank.db")

	store, err := NewStore(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer store.Close()

	app := &App{store: store, sessions: NewSessionManager()}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/signup", app.handleSignup)
	mux.HandleFunc("POST /api/login", app.handleLogin)
	mux.HandleFunc("POST /api/logout", app.handleLogout)
	mux.HandleFunc("GET /api/me", app.requireAuth(app.handleMe))
	mux.HandleFunc("GET /api/balance", app.requireAuth(app.handleBalance))
	mux.HandleFunc("POST /api/deposit", app.requireAuth(app.handleDeposit))
	mux.HandleFunc("POST /api/withdraw", app.requireAuth(app.handleWithdraw))
	mux.HandleFunc("GET /api/transactions", app.requireAuth(app.handleTransactions))
	mux.Handle("/", http.FileServer(http.Dir(getenv("STATIC_DIR", "static"))))

	log.Printf("MIeBank listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, logRequests(mux)))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
