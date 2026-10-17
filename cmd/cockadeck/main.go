package main

// point d'entrée de l'app avec la conf de router.go
import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"cockadeck/internal/catalog"
	"cockadeck/internal/config"
	"cockadeck/internal/scryfall"
	"cockadeck/internal/server"
	"cockadeck/internal/store"
	"cockadeck/internal/store/sqlcgen"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var handler slog.Handler
	if cfg.Prod() {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))

	db, err := store.Open(context.Background(), cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	q := sqlcgen.New(db)
	svc := catalog.New(q, db, scryfall.NewClient(cfg.ScryfallURL))
	srv := server.New(cfg, q, svc)
	slog.Info("listening", "addr", cfg.Addr, "env", cfg.Env)
	return http.ListenAndServe(cfg.Addr, srv)
}
