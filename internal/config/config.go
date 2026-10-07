package config

import (
	"fmt"
	"os"
)

type Config struct {
	Addr        string
	DBPath      string
	JWTSecret   string
	Env         string
	ScryfallURL string
}

func (c Config) Prod() bool { return c.Env == "prod" }

func Load() (Config, error) {
	cfg := Config{
		Addr:        envOr("COCKADECK_ADDR", ":8080"),
		DBPath:      envOr("COCKADECK_DB_PATH", "./cockadeck.db"),
		JWTSecret:   os.Getenv("COCKADECK_JWT_SECRET"),
		Env:         envOr("COCKADECK_ENV", "dev"),
		ScryfallURL: envOr("COCKADECK_SCRYFALL_URL", "https://api.scryfall.com"),
	}
	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("COCKADECK_JWT_SECRET is required")
	}
	if cfg.Env != "dev" && cfg.Env != "prod" {
		return cfg, fmt.Errorf(`COCKADECK_ENV must be "dev" or "prod", got %q`, cfg.Env)
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
