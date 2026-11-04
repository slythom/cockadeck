package views

import (
	"context"
	"strings"
	"testing"
)

func TestManaTokens(t *testing.T) {
	cases := []struct {
		cost string
		want []manaToken
	}{
		{"{2}{R}", []manaToken{{"2", true}, {"R", true}}},
		{"{W/U}{W/U}", []manaToken{{"W/U", true}, {"W/U", true}}},
		{"{1}{R} // {R}", []manaToken{{"1", true}, {"R", true}, {" // ", false}, {"R", true}}},
		{"{X}{U}{U}", []manaToken{{"X", true}, {"U", true}, {"U", true}}},
	}
	for _, c := range cases {
		got := manaTokens(c.cost)
		if len(got) != len(c.want) {
			t.Fatalf("%s: got %d tokens, want %d", c.cost, len(got), len(c.want))
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: token %d = %+v, want %+v", c.cost, i, got[i], c.want[i])
			}
		}
	}
}

func TestManaGlyphID(t *testing.T) {
	for tok, want := range map[string]string{
		"W": "ms-w", "U": "ms-u", "B": "ms-b", "R": "ms-r", "G": "ms-g",
		"C": "ms-c", "S": "ms-s", "T": "ms-t",
		"W/U": "ms-w-u", "G/W": "ms-g-w", "B/R": "ms-b-r",
		"W/P": "ms-w-p", "B/P": "ms-b-p",
		"2/W": "ms-2-w", "2/G": "ms-2-g",
		"w":   "ms-w", // case-insensitive
	} {
		if id, ok := manaGlyphID(tok); !ok || id != want {
			t.Errorf("manaGlyphID(%q) = %q, %v; want %q", tok, id, ok, want)
		}
	}
	for _, tok := range []string{"2", "10", "X", "Y", "∞", "W/U/P", "H"} {
		if id, ok := manaGlyphID(tok); ok {
			t.Errorf("manaGlyphID(%q) = %q; want no glyph", tok, id)
		}
	}
}

func TestManaCostRender(t *testing.T) {
	var b strings.Builder
	if err := ManaCost("{1}{W/U}{C}").Render(context.Background(), &b); err != nil {
		t.Fatal(err)
	}
	html := b.String()
	for _, frag := range []string{`title="{1}{W/U}{C}"`, `ms ms-generic`, `mana.svg#ms-w-u`, `mana.svg#ms-c`} {
		if !strings.Contains(html, frag) {
			t.Errorf("rendered html missing %q:\n%s", frag, html)
		}
	}
}
