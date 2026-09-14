package handlers

import (
	"cockadeck/views"
	"net/http"
)

// code qui doit être exécuté par les routes définies dans router.go

func ParseForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	setCode := r.FormValue("setCode")
	collectorNumber := r.FormValue("collectorNumber")

	w.Header().Set("Content-Type", "text/html") // Later: make a proper Layout.templ with correct html tags
	views.SearchCard(setCode, collectorNumber).Render(r.Context(), w)
}
