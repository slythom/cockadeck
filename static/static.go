// Package static embeds the vendored frontend assets.
package static

import "embed"

//go:embed app.css app.js htmx.min.js favicon.svg mana.svg fonts
var FS embed.FS
