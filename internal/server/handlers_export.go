package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"cockadeck/internal/catalog"
	"cockadeck/internal/export"
)

func (s *Server) exportDeck(w http.ResponseWriter, r *http.Request) {
	deck, ok := s.deckForUser(w, r)
	if !ok {
		return
	}
	rows, err := s.cards.DeckCards(r.Context(), deck.ID, catalog.CardFilters{})
	if err != nil {
		slog.Error("export deck: list cards", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	body, err := export.Cockatrice(deck, rows)
	if err != nil {
		slog.Error("export deck", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%q", export.Slug(deck.Name)+".cod"))
	w.Write(body)
}
