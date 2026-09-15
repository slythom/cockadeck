package handlers

import (
	"cockadeck/views"
	"log/slog"
	"net/http"
	"net/url"
)

// code qui doit être exécuté par les routes définies dans router.go

func ParseForm(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handler start........")
	if r.Method == http.MethodPost {

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		setCode := r.FormValue("setCode")
		collectorNumber := r.FormValue("collectorNumber")

		slog.Info("POST / setCode:" + setCode + " collectorNumber:" + collectorNumber)

		http.Redirect(w, r, "/?setCode="+url.QueryEscape(setCode)+
			"&collectorNumber="+url.QueryEscape(collectorNumber),
			http.StatusSeeOther)
		return
	}

	setCode := r.URL.Query().Get("setCode")
	collectorNumber := r.URL.Query().Get("collectorNumber")
	slog.Info("GET / setCode:" + setCode + " collectorNumber:" + collectorNumber)

	w.Header().Set("Content-Type", "text/html") // Later: make a proper Layout.templ with correct html tags
	views.SearchCard(setCode, collectorNumber).Render(r.Context(), w)

}
