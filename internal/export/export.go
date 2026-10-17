// Package export writes deck and collection files (.cod XML, CSV).
package export

import (
	"encoding/xml"
	"fmt"
	"strings"
	"unicode"

	"cockadeck/internal/store/sqlcgen"
)

type codCard struct {
	Number int64  `xml:"number,attr"`
	Name   string `xml:"name,attr"`
}

type codZone struct {
	Name  string    `xml:"name,attr"`
	Cards []codCard `xml:"card"`
}

// add merges same-name cards into one quantity — .cod is name-keyed, so
// two printings (or a commander/main split) of a card become a single row.
func (z *codZone) add(name string, qty int64) {
	for i := range z.Cards {
		if z.Cards[i].Name == name {
			z.Cards[i].Number += qty
			return
		}
	}
	z.Cards = append(z.Cards, codCard{Number: qty, Name: name})
}

type codDeck struct {
	XMLName  xml.Name  `xml:"cockatrice_deck"`
	Version  string    `xml:"version,attr"`
	DeckName string    `xml:"deckname"`
	Comments string    `xml:"comments"`
	Zones    []codZone `xml:"zone"`
}

// Cockatrice renders a deck as Cockatrice .cod XML. Zones "main" and
// "commander" merge into <zone name="main">; "sideboard" becomes "side" (D12).
func Cockatrice(deck sqlcgen.Deck, rows []sqlcgen.ListDeckCardsRow) ([]byte, error) {
	main := codZone{Name: "main"}
	side := codZone{Name: "side"}
	for _, r := range rows {
		if r.Zone == "sideboard" {
			side.add(r.Name, r.Quantity)
		} else {
			main.add(r.Name, r.Quantity)
		}
	}
	d := codDeck{
		Version:  "1",
		DeckName: deck.Name,
		Comments: deck.Description,
		Zones:    []codZone{main, side},
	}
	out, err := xml.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal .cod: %w", err)
	}
	return append([]byte(xml.Header), out...), nil
}

// Slug makes a filename-safe deck name.
func Slug(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if dash {
				b.WriteByte('-')
				dash = false
			}
			b.WriteRune(r)
		} else {
			dash = b.Len() > 0
		}
	}
	if s := b.String(); s != "" {
		return s
	}
	return "deck"
}
