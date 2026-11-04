package server

import (
	"net/http"
	"strings"
)

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// wantsFragment reports whether this htmx request targets a partial swap.
// htmx v4 sends HX-Target as "tagName#id" (e.g. "div#deck-cards"), so a
// suffix match on "#id" identifies the fragment the client expects.
func wantsFragment(r *http.Request, target string) bool {
	return isHTMX(r) && strings.HasSuffix(r.Header.Get("HX-Target"), target)
}

func redirect(w http.ResponseWriter, r *http.Request, url string) {
	if isHTMX(r) {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}
