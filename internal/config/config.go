// Package config reads server settings from environment variables.
package config

import "os"

type Config struct {
	Addr      string
	DBPath    string
	StaticDir string
}

func Load() Config {
	return Config{
		Addr:      getenv("ADDR", ":8080"),
		DBPath:    getenv("DB_PATH", "data/miebank.db"),
		StaticDir: getenv("STATIC_DIR", "frontend/dist"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
