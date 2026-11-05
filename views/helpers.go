package views

import (
	"hash/fnv"
	"io"
	"net/url"
	"strconv"

	"cockadeck/internal/store/sqlcgen"
	"cockadeck/static"

	"github.com/a-h/templ"
)

// staticVer fingerprints the embedded assets so stylesheet/script URLs change
// whenever the binary is rebuilt — the /static/ cache is max-age=1h.
var staticVer = func() string {
	h := fnv.New64a()
	for _, name := range []string{"app.css", "app.js", "htmx.min.js"} {
		if f, err := static.FS.Open(name); err == nil {
			io.Copy(h, f)
			f.Close()
		}
	}
	return strconv.FormatUint(h.Sum64(), 36)
}()

func staticURL(name string) string {
	return "/static/" + name + "?v=" + staticVer
}

// cardListSwap is the shared contract for deck-row controls: submit the
// current filters and morph the re-rendered card list back into place.
var cardListSwap = templ.Attributes{
	"hx-include": "#deck-filters",
	"hx-target":  "#deck-cards",
	"hx-swap":    "innerMorph",
}

type DeckFilters struct {
	Set    string
	Rarity string
	CMC    string
}

var rarities = []string{"common", "uncommon", "rare", "mythic"}

func deckURL(id int64) string {
	return "/decks/" + strconv.FormatInt(id, 10)
}

func deckCardsURL(deckID int64) string {
	return deckURL(deckID) + "/cards"
}

func deckCardRowURL(deckID int64, cardID, zone string) string {
	return deckCardsURL(deckID) + "/" + cardID + "?zone=" + url.QueryEscape(zone)
}

// zoneOrder lists display order; commander decks lead with the commander zone.
func zoneOrder(format string) []string {
	if format == "commander" {
		return []string{"commander", "main", "sideboard"}
	}
	return []string{"main", "sideboard"}
}

func zoneRows(rows []sqlcgen.ListDeckCardsRow, zone string) []sqlcgen.ListDeckCardsRow {
	var out []sqlcgen.ListDeckCardsRow
	for _, r := range rows {
		if r.Zone == zone {
			out = append(out, r)
		}
	}
	return out
}

func zoneTotal(rows []sqlcgen.ListDeckCardsRow, zone string) int64 {
	var n int64
	for _, r := range rows {
		if r.Zone == zone {
			n += r.Quantity
		}
	}
	return n
}

// shortDate trims a SQLite datetime ("YYYY-MM-DD HH:MM:SS") to its date part.
func shortDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func totalCards(rows []sqlcgen.ListDeckCardsRow) int64 {
	var n int64
	for _, r := range rows {
		n += r.Quantity
	}
	return n
}
