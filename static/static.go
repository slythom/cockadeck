// Package static embeds the vendored frontend assets.
package static

import "embed"

//go:embed app.css app.js htmx.min.js favicon.svg
var FS embed.FS
