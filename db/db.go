// Package db exposes the embedded goose migrations.
package db

import "embed"

//go:embed migrations
var Migrations embed.FS
