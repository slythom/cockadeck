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

// maxSearchResults caps how many name-search rows are rendered; Scryfall
// pages up to 175 cards which is too heavy a fragment at this scale.
const maxSearchResults = 50

// searchOutcome is the view-model for both search modes: a single printing
// (set+num) or a list of matches (name).
type searchOutcome struct {
	card   *sqlcgen.Card
	cards  []sqlcgen.Card
	total  int
	decks  []sqlcgen.ListDecksRow
	errMsg string
}

func (s *Server) searchPage(w http.ResponseWriter, r *http.Request) {
	set, num, name := searchParams(r)
	var res searchOutcome
	if set != "" || num != "" || name != "" {
		res = s.runSearch(r, set, num, name)
	}
	views.SearchPage(set, num, name, res.card, res.cards, res.total, res.decks, res.errMsg).Render(r.Context(), w)
}

func (s *Server) searchResults(w http.ResponseWriter, r *http.Request) {
	set, num, name := searchParams(r)
	w.Header().Set("HX-Push-Url", searchURL(set, num, name))
	res := s.runSearch(r, set, num, name)
	views.SearchResults(res.card, res.cards, res.total, res.decks, res.errMsg).Render(r.Context(), w)
}

func searchParams(r *http.Request) (set, num, name string) {
	return r.URL.Query().Get("set"), r.URL.Query().Get("num"), r.URL.Query().Get("name")
}

// searchURL rebuilds the shareable /search URL from whichever params are set.
func searchURL(set, num, name string) string {
	var q strings.Builder
	q.WriteString("/search")
	sep := "?"
	for _, kv := range [][2]string{{"set", set}, {"num", num}, {"name", name}} {
		if v := strings.TrimSpace(kv[1]); v != "" {
			q.WriteString(sep + kv[0] + "=" + url.QueryEscape(v))
			sep = "&"
		}
	}
	return q.String()
}

// runSearch dispatches on the filled fields: a name wins, otherwise set+num
// is required. Decks are loaded once when there is anything to add.
func (s *Server) runSearch(r *http.Request, set, num, name string) searchOutcome {
	var out searchOutcome
	switch name = strings.TrimSpace(name); {
	case name != "":
		out.cards, out.total, out.errMsg = s.searchByName(r.Context(), name, strings.TrimSpace(set))
	default:
		out.card, out.errMsg = s.lookupCard(r.Context(), set, num)
	}
	if out.errMsg == "" && (out.card != nil || len(out.cards) > 0) {
		out.decks, _ = s.cards.ListDecks(r.Context(), userID(r))
	}
	if len(out.cards) > maxSearchResults {
		out.cards = out.cards[:maxSearchResults]
	}
	return out
}

// lookupCard returns (card, "") on success or (nil, user-facing message) on failure.
func (s *Server) lookupCard(ctx context.Context, set, num string) (*sqlcgen.Card, string) {
	set = strings.TrimSpace(set)
	num = strings.TrimSpace(num)
	if !setCodeRe.MatchString(set) || num == "" || len(num) > 15 {
		return nil, `Enter a set code (e.g. "otj") and collector number (e.g. "10"), or search by name.`
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

// searchByName is the name-mode counterpart of lookupCard: (cards, total, msg).
func (s *Server) searchByName(ctx context.Context, name, set string) ([]sqlcgen.Card, int, string) {
	if len(name) > 200 {
		return nil, 0, "That search is too long — keep it under 200 characters."
	}
	cards, total, err := s.cards.CardsByName(ctx, name, set)
	switch {
	case err == nil:
		return cards, total, ""
	case errors.Is(err, scryfall.ErrNotFound):
		return nil, 0, fmt.Sprintf("No cards match %q.", name)
	case errors.Is(err, scryfall.ErrInvalidInput):
		return nil, 0, "Scryfall rejected that search — check the card name."
	default:
		slog.Error("scryfall name search", "name", name, "err", err)
		return nil, 0, "Scryfall is unavailable right now — try again shortly."
	}
}
