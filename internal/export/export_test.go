package export

import (
	"os"
	"strings"
	"testing"

	"cockadeck/internal/store/sqlcgen"
)

func TestCockatrice(t *testing.T) {
	deck := sqlcgen.Deck{ID: 1, Name: "Gruul Smash", Description: "fast & loose", Format: "commander"}
	rows := []sqlcgen.ListDeckCardsRow{
		{CardID: "a", Zone: "commander", Quantity: 1, Name: "Zilortha, Strength Incarnate"},
		{CardID: "b", Zone: "main", Quantity: 4, Name: "Lightning Bolt"},
		{CardID: "c", Zone: "main", Quantity: 20, Name: "Forest"},
		{CardID: "c2", Zone: "main", Quantity: 4, Name: "Forest"},
		{CardID: "d", Zone: "sideboard", Quantity: 2, Name: "Veil of Summer"},
	}
	got, err := Cockatrice(deck, rows)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/deck.cod")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(got)) != strings.TrimSpace(string(want)) {
		t.Fatalf("mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Gruul Smash":     "gruul-smash",
		"  My  Deck!! ":   "my-deck",
		"déjà vu":         "déjà-vu",
		"":                "deck",
		"---":             "deck",
		"Atraxa's +1/+1s": "atraxa-s-1-1s",
	} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}
