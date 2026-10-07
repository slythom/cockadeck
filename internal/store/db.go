package store

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	dbfs "cockadeck/db"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, path string) (*sql.DB, error) {
	q := url.Values{}
	for _, p := range []string{
		"busy_timeout(5000)",
		"journal_mode(WAL)",
		"foreign_keys(1)",
		"synchronous(NORMAL)",
	} {
		q.Add("_pragma", p)
	}
	db, err := sql.Open("sqlite", "file:"+path+"?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(4)

	goose.SetBaseFS(dbfs.Migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}
