package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"cockadeck/internal/auth"
	"cockadeck/internal/catalog"
	"cockadeck/internal/store/sqlcgen"
	"cockadeck/views"
)

func userID(r *http.Request) int64 {
	uid, _ := auth.UserIDFromContext(r.Context())
	return uid
}

func parseDeckID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// deckForUser resolves /decks/{id} and enforces ownership. On failure it writes
// the response (404/500) and returns ok=false.
func (s *Server) deckForUser(w http.ResponseWriter, r *http.Request) (sqlcgen.Deck, bool) {
	id, err := parseDeckID(r)
	if err != nil {
		http.NotFound(w, r)
		return sqlcgen.Deck{}, false
	}
	deck, err := s.cards.GetDeck(r.Context(), id, userID(r))
	switch {
	case errors.Is(err, catalog.ErrNotFound):
		http.NotFound(w, r)
		return sqlcgen.Deck{}, false
	case err != nil:
		slog.Error("get deck", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return sqlcgen.Deck{}, false
	}
	return deck, true
}

func parseFilters(r *http.Request) (catalog.CardFilters, views.DeckFilters) {
	set := strings.TrimSpace(r.FormValue("set"))
	rarity := r.FormValue("rarity")
	cmcStr := r.FormValue("cmc")
	f := catalog.CardFilters{Set: set, Rarity: rarity}
	if cmcStr != "" {
		if v, err := strconv.ParseFloat(cmcStr, 64); err == nil {
			f.CMC = &v
		}
	}
	return f, views.DeckFilters{Set: set, Rarity: rarity, CMC: cmcStr}
}

func (s *Server) decksPage(w http.ResponseWriter, r *http.Request) {
	s.renderDecksPage(w, r, "", http.StatusOK)
}

func (s *Server) renderDecksPage(w http.ResponseWriter, r *http.Request, errMsg string, code int) {
	decks, err := s.cards.ListDecks(r.Context(), userID(r))
	if err != nil {
		slog.Error("list decks", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if code != http.StatusOK {
		w.WriteHeader(code)
	}
	views.DecksPage(decks, errMsg).Render(r.Context(), w)
}

func (s *Server) createDeck(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	deck, err := s.cards.CreateDeck(r.Context(), userID(r),
		r.FormValue("name"), r.FormValue("description"), r.FormValue("format"))
	switch {
	case err == nil:
		redirect(w, r, fmt.Sprintf("/decks/%d", deck.ID))
	case errors.Is(err, catalog.ErrValidation):
		s.renderDecksPage(w, r, err.Error(), http.StatusUnprocessableEntity)
	default:
		slog.Error("create deck", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (s *Server) deckPage(w http.ResponseWriter, r *http.Request) {
	deck, ok := s.deckForUser(w, r)
	if !ok {
		return
	}
	f, vf := parseFilters(r)
	rows, err := s.cards.DeckCards(r.Context(), deck.ID, f)
	if err != nil {
		slog.Error("list deck cards", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	views.DeckPage(deck, rows, vf, "").Render(r.Context(), w)
}

func (s *Server) updateDeck(w http.ResponseWriter, r *http.Request) {
	deck, ok := s.deckForUser(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	err := s.cards.UpdateDeck(r.Context(), deck.ID, userID(r),
		r.FormValue("name"), r.FormValue("description"), r.FormValue("format"))
	switch {
	case err == nil:
		redirect(w, r, fmt.Sprintf("/decks/%d", deck.ID))
	case errors.Is(err, catalog.ErrValidation):
		f, vf := parseFilters(r)
		rows, _ := s.cards.DeckCards(r.Context(), deck.ID, f)
		w.WriteHeader(http.StatusUnprocessableEntity)
		views.DeckPage(deck, rows, vf, err.Error()).Render(r.Context(), w)
	default:
		slog.Error("update deck", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (s *Server) deleteDeck(w http.ResponseWriter, r *http.Request) {
	deck, ok := s.deckForUser(w, r)
	if !ok {
		return
	}
	if err := s.cards.DeleteDeck(r.Context(), deck.ID, userID(r)); err != nil {
		slog.Error("delete deck", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	redirect(w, r, "/decks")
}

func (s *Server) addDeckCard(w http.ResponseWriter, r *http.Request) {
	deckID, err := parseDeckID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	cardID := r.FormValue("card_id")
	zone := r.FormValue("zone")
	if zone == "" {
		zone = "main"
	}
	qty, err := strconv.ParseInt(r.FormValue("quantity"), 10, 64)
	if err != nil || qty < 1 {
		qty = 1
	}

	err = s.cards.AddDeckCard(r.Context(), deckID, userID(r), cardID, zone, qty)
	switch {
	case err == nil:
		views.AddCardStatus(true, "added").Render(r.Context(), w)
	case errors.Is(err, catalog.ErrValidation):
		views.AddCardStatus(false, err.Error()).Render(r.Context(), w)
	case errors.Is(err, catalog.ErrNotFound):
		views.AddCardStatus(false, "deck not found").Render(r.Context(), w)
	default:
		slog.Error("add deck card", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// renderDeckCards re-renders the #deck-cards fragment honoring any filters that
// rode along on the request (controls hx-include #deck-filters).
func (s *Server) renderDeckCards(w http.ResponseWriter, r *http.Request, deck sqlcgen.Deck, errMsg string) {
	f, _ := parseFilters(r)
	rows, err := s.cards.DeckCards(r.Context(), deck.ID, f)
	if err != nil {
		slog.Error("list deck cards", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	views.DeckCardList(deck, rows, errMsg).Render(r.Context(), w)
}

func (s *Server) updateDeckCard(w http.ResponseWriter, r *http.Request) {
	deck, ok := s.deckForUser(w, r)
	if !ok {
		return
	}
	cardID := r.PathValue("cardID")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	zone := r.FormValue("zone")
	if zone == "" {
		http.Error(w, "zone required", http.StatusBadRequest)
		return
	}
	var quantity *int64
	if v := r.FormValue("quantity"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			quantity = &n
		}
	}
	newZone := r.FormValue("new_zone")

	err := s.cards.UpdateDeckCard(r.Context(), deck.ID, userID(r), cardID, zone, quantity, newZone)
	var errMsg string
	switch {
	case err == nil:
	case errors.Is(err, catalog.ErrValidation), errors.Is(err, catalog.ErrNotFound):
		errMsg = err.Error()
	default:
		slog.Error("update deck card", "err", err)
		errMsg = "update failed"
	}
	s.renderDeckCards(w, r, deck, errMsg)
}

func (s *Server) deleteDeckCard(w http.ResponseWriter, r *http.Request) {
	deck, ok := s.deckForUser(w, r)
	if !ok {
		return
	}
	cardID := r.PathValue("cardID")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	zone := r.FormValue("zone")
	if zone == "" {
		http.Error(w, "zone required", http.StatusBadRequest)
		return
	}
	var errMsg string
	if err := s.cards.RemoveDeckCard(r.Context(), deck.ID, userID(r), cardID, zone); err != nil {
		slog.Error("delete deck card", "err", err)
		errMsg = "remove failed"
	}
	s.renderDeckCards(w, r, deck, errMsg)
}
