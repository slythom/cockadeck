// Package static embeds the vendored frontend assets.
package static

import "embed"

//go:embed app.css htmx.min.js
var FS embed.FS
