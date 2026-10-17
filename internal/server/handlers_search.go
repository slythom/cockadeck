package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"cockadeck/internal/scryfall"
	"cockadeck/internal/store/sqlcgen"
	"cockadeck/views"
)

var setCodeRe = regexp.MustCompile(`^[a-zA-Z0-9]{1,8}$`)

func (s *Server) searchPage(w http.ResponseWriter, r *http.Request) {
	set := r.URL.Query().Get("set")
	num := r.URL.Query().Get("num")
	var card *sqlcgen.Card
	var errMsg string
	if set != "" || num != "" {
		card, errMsg = s.lookupCard(r.Context(), set, num)
	}
	var decks []sqlcgen.ListDecksRow
	if card != nil {
		decks, _ = s.cards.ListDecks(r.Context(), userID(r))
	}
	views.SearchPage(set, num, card, decks, errMsg).Render(r.Context(), w)
}

func (s *Server) searchResults(w http.ResponseWriter, r *http.Request) {
	set := r.URL.Query().Get("set")
	num := r.URL.Query().Get("num")
	card, errMsg := s.lookupCard(r.Context(), set, num)
	w.Header().Set("HX-Push-Url", "/search?set="+url.QueryEscape(set)+"&num="+url.QueryEscape(num))
	var decks []sqlcgen.ListDecksRow
	if card != nil {
		decks, _ = s.cards.ListDecks(r.Context(), userID(r))
	}
	views.SearchResults(card, decks, errMsg).Render(r.Context(), w)
}

// lookupCard returns (card, "") on success or (nil, user-facing message) on failure.
func (s *Server) lookupCard(ctx context.Context, set, num string) (*sqlcgen.Card, string) {
	set = strings.TrimSpace(set)
	num = strings.TrimSpace(num)
	if !setCodeRe.MatchString(set) || num == "" || len(num) > 15 {
		return nil, `Enter a set code (e.g. "otj") and collector number (e.g. "10").`
	}

	c, err := s.cards.CardBySetNumber(ctx, set, num)
	switch {
	case err == nil:
		return &c, ""
	case errors.Is(err, scryfall.ErrNotFound):
		return nil, fmt.Sprintf("No card found for %s #%s.", strings.ToUpper(set), num)
	case errors.Is(err, scryfall.ErrInvalidInput):
		return nil, "Scryfall rejected that lookup — check the set code and collector number."
	default:
		slog.Error("scryfall lookup", "set", set, "num", num, "err", err)
		return nil, "Scryfall is unavailable right now — try again shortly."
	}
}
