package router

import (
	"net/http"

	"cockadeck/handlers"
)

// Oui. Un dossier router ou routes sert généralement à regrouper la définition des routes HTTP,
// c'est-à-dire la correspondance entre une URL et le code qui doit être exécuté.

func Setup() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handlers.ParseForm)
	mux.HandleFunc("POST /", handlers.ParseForm)

	return mux
}
