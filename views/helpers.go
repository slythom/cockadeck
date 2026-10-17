package views

import (
	"net/url"
	"strconv"

	"cockadeck/internal/store/sqlcgen"
)

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

func totalCards(rows []sqlcgen.ListDeckCardsRow) int64 {
	var n int64
	for _, r := range rows {
		n += r.Quantity
	}
	return n
}

func cardCount(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	}
	return 0
}
