package views

import "strings"

type manaToken struct {
	// Sym is the symbol text — "2", "W", "W/U" for braced tokens, or the
	// verbatim text for unbraced segments (" // " between faces of a DFC).
	Sym    string
	Braced bool
}

// manaTokens splits a Scryfall mana cost like "{2}{R}{W/U}" into its
// individual symbols. Unbraced segments (e.g. "//" in "{1}{R} // {R}")
// are kept verbatim and marked Braced=false.
func manaTokens(cost string) []manaToken {
	var toks []manaToken
	for len(cost) > 0 {
		open := strings.IndexByte(cost, '{')
		if open < 0 {
			toks = append(toks, manaToken{Sym: cost})
			break
		}
		if open > 0 {
			toks = append(toks, manaToken{Sym: cost[:open]})
		}
		close := strings.IndexByte(cost[open:], '}')
		if close < 0 {
			toks = append(toks, manaToken{Sym: cost[open:]})
			break
		}
		toks = append(toks, manaToken{Sym: cost[open+1 : open+close], Braced: true})
		cost = cost[open+close+1:]
	}
	return toks
}

// manaGlyphID maps a mana symbol to the id of its <symbol> in mana.svg.
// Unrecognized tokens ("2", "X", "W/U/P", ...) have no glyph and are
// rendered as a generic circle with text instead.
func manaGlyphID(tok string) (string, bool) {
	t := strings.ToUpper(strings.TrimSpace(tok))
	switch t {
	case "W", "U", "B", "R", "G", "C", "S", "T":
		return "ms-" + strings.ToLower(t), true
	}
	parts := strings.Split(t, "/")
	if len(parts) != 2 {
		return "", false
	}
	a, b := parts[0], parts[1]
	switch {
	case isManaColor(a) && isManaColor(b), isManaColor(a) && b == "P", a == "2" && isManaColor(b):
		return "ms-" + strings.ToLower(a) + "-" + strings.ToLower(b), true
	}
	return "", false
}

func isManaColor(s string) bool {
	switch s {
	case "W", "U", "B", "R", "G":
		return true
	}
	return false
}

// genericTextClass shrinks the label inside the generic-mana circle for
// multi-character tokens ("10", "∞", "W/U/P").
func genericTextClass(tok string) string {
	switch n := len([]rune(tok)); {
	case n >= 3:
		return "ms-txt ms-txt-3"
	case n == 2:
		return "ms-txt ms-txt-2"
	}
	return "ms-txt"
}
