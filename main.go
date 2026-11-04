package main

import (
	"log"
	"net/http"

	"miebank/internal/api"
	"miebank/internal/config"
	"miebank/internal/session"
	"miebank/internal/store"
	"miebank/internal/web"
)

func main() {
	cfg := config.Load()

	db, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	server := api.NewServer(db, session.NewManager())

	mux := http.NewServeMux()
	server.Register(mux)
	mux.Handle("/", web.SPA(cfg.StaticDir))

	log.Printf("MIeBank listening on %s (frontend: %s)", cfg.Addr, cfg.StaticDir)
	log.Fatal(http.ListenAndServe(cfg.Addr, api.LogRequests(mux)))
}
