package server_test

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"cockadeck/internal/store/sqlcgen"
)

// authedForm sends a form request (any method) with session cookie + Origin.
func authedForm(t *testing.T, srv *httptest.Server, client *http.Client, cookie *http.Cookie, method, path string, form url.Values) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", srv.URL)
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func seedCard(t *testing.T, q *sqlcgen.Queries, id, name string) {
	t.Helper()
	seedCardFull(t, q, id, name, "", "")
}

func seedCardFull(t *testing.T, q *sqlcgen.Queries, id, name, manaCost, colorIdentity string) {
	t.Helper()
	_, err := q.UpsertCard(context.Background(), sqlcgen.UpsertCardParams{
		ID: id, OracleID: "o-" + id, Name: name, SetCode: "otj",
		SetName: "Outlaws of Thunder Junction", CollectorNumber: "1",
		Rarity: "common", ManaCost: sql.NullString{String: manaCost, Valid: manaCost != ""},
		ColorIdentity: colorIdentity,
		Cmc:           sql.NullFloat64{Float64: 2, Valid: true},
		ImageUri:      "https://img/x.jpg", Raw: "{}",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func body(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}

func TestDeckFlow(t *testing.T) {
	h, q := newTestServer(t, "http://127.0.0.1:1")
	srv := httptest.NewServer(h)
	defer srv.Close()
	client := noRedirectClient()
	cookie := registerAndCookie(t, srv, client)

	// Create a commander deck → redirect to its page.
	resp := authedForm(t, srv, client, cookie, http.MethodPost, "/decks", url.Values{
		"name": {"Test Deck"}, "format": {"commander"}, "description": {"desc"},
	})
	loc := resp.Header.Get("Location")
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther || !strings.HasPrefix(loc, "/decks/") {
		t.Fatalf("create deck: %d %q", resp.StatusCode, loc)
	}
	deckPath := loc

	seedCard(t, q, "card-1", "Lightning Greaves")

	// Add card to main zone.
	resp = authedForm(t, srv, client, cookie, http.MethodPost, deckPath+"/cards", url.Values{
		"card_id": {"card-1"}, "zone": {"main"}, "quantity": {"2"},
	})
	if b := body(t, resp); !strings.Contains(b, "added") {
		t.Fatalf("add card: %s", b)
	}

	// Deck page shows the card with quantity 2.
	resp = authedGet(t, srv, client, cookie, deckPath)
	if b := body(t, resp); !strings.Contains(b, "Lightning Greaves") || !strings.Contains(b, `class="qty">2`) {
		t.Fatalf("deck page missing card/qty")
	}

	// Quantity up to 5.
	resp = authedForm(t, srv, client, cookie, http.MethodPatch, deckPath+"/cards/card-1?zone=main", url.Values{
		"quantity": {"5"},
	})
	if b := body(t, resp); !strings.Contains(b, `class="qty">5`) {
		t.Fatalf("qty update failed: %.200s", b)
	}

	// Move to commander zone.
	resp = authedForm(t, srv, client, cookie, http.MethodPatch, deckPath+"/cards/card-1?zone=main", url.Values{
		"new_zone": {"commander"},
	})
	if b := body(t, resp); !strings.Contains(b, "badge-commander") {
		t.Fatalf("zone move failed: %.200s", b)
	}

	// Filter by rarity: common matches, mythic doesn't.
	resp = authedGet(t, srv, client, cookie, deckPath+"?rarity=mythic")
	if b := body(t, resp); strings.Contains(b, "Lightning Greaves") {
		t.Fatal("rarity filter did not exclude card")
	}
	resp = authedGet(t, srv, client, cookie, deckPath+"?rarity=common")
	if b := body(t, resp); !strings.Contains(b, "Lightning Greaves") {
		t.Fatal("rarity filter excluded matching card")
	}

	// Export: commander merges into main, headers set for download.
	resp = authedForm(t, srv, client, cookie, http.MethodPost, deckPath+"/cards", url.Values{
		"card_id": {"card-1"}, "zone": {"sideboard"}, "quantity": {"2"},
	})
	resp.Body.Close()
	resp = authedGet(t, srv, client, cookie, deckPath+"/export.cod")
	cd := resp.Header.Get("Content-Disposition")
	if ct := resp.Header.Get("Content-Type"); ct != "application/octet-stream" {
		t.Fatalf("export content-type = %q", ct)
	}
	if !strings.Contains(cd, "test-deck.cod") {
		t.Fatalf("export content-disposition = %q", cd)
	}
	exportBody := body(t, resp)
	if !strings.Contains(exportBody, `<deckname>Test Deck</deckname>`) ||
		!strings.Contains(exportBody, `<zone name="main">`) ||
		!strings.Contains(exportBody, `<zone name="side">`) ||
		!strings.Contains(exportBody, `name="Lightning Greaves"`) {
		t.Fatalf("export body wrong:\n%s", exportBody)
	}

	// Remove both remaining rows (commander + sideboard).
	resp = authedForm(t, srv, client, cookie, http.MethodDelete, deckPath+"/cards/card-1?zone=commander", nil)
	resp.Body.Close()
	resp = authedForm(t, srv, client, cookie, http.MethodDelete, deckPath+"/cards/card-1?zone=sideboard", nil)
	if b := body(t, resp); strings.Contains(b, "Lightning Greaves") {
		t.Fatal("delete failed")
	}
}

func TestDeckZoneRule(t *testing.T) {
	h, q := newTestServer(t, "http://127.0.0.1:1")
	srv := httptest.NewServer(h)
	defer srv.Close()
	client := noRedirectClient()
	cookie := registerAndCookie(t, srv, client)

	resp := authedForm(t, srv, client, cookie, http.MethodPost, "/decks", url.Values{
		"name": {"Forty"}, "format": {"40-card"},
	})
	deckPath := resp.Header.Get("Location")
	resp.Body.Close()

	seedCard(t, q, "card-1", "Bolt")

	resp = authedForm(t, srv, client, cookie, http.MethodPost, deckPath+"/cards", url.Values{
		"card_id": {"card-1"}, "zone": {"commander"}, "quantity": {"1"},
	})
	if b := body(t, resp); !strings.Contains(b, "commander zone only on commander-format") {
		t.Fatalf("expected zone rule error, got: %s", b)
	}
}

func TestDeckPageHTMXFragment(t *testing.T) {
	h, q := newTestServer(t, "http://127.0.0.1:1")
	srv := httptest.NewServer(h)
	defer srv.Close()
	client := noRedirectClient()
	cookie := registerAndCookie(t, srv, client)

	resp := authedForm(t, srv, client, cookie, http.MethodPost, "/decks", url.Values{
		"name": {"Frag"}, "format": {"40-card"},
	})
	deckPath := resp.Header.Get("Location")
	resp.Body.Close()

	seedCard(t, q, "card-1", "Bolt")
	resp = authedForm(t, srv, client, cookie, http.MethodPost, deckPath+"/cards", url.Values{
		"card_id": {"card-1"}, "zone": {"main"}, "quantity": {"1"},
	})
	resp.Body.Close()

	// htmx request targeting #deck-cards gets only the list fragment.
	req, err := http.NewRequest(http.MethodGet, srv.URL+deckPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(cookie)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "div#deck-cards")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	b := body(t, resp)
	if !strings.Contains(b, "Bolt") {
		t.Fatalf("fragment missing card row: %.200s", b)
	}
	for _, frag := range []string{`class="topbar"`, "deck-head", `id="deck-filters"`, `id="deck-cards"`} {
		if strings.Contains(b, frag) {
			t.Fatalf("fragment contains full-page markup %q", frag)
		}
	}

	// Plain GET still gets the full page.
	resp = authedGet(t, srv, client, cookie, deckPath)
	b = body(t, resp)
	if !strings.Contains(b, `class="topbar"`) || !strings.Contains(b, `id="deck-cards"`) {
		t.Fatal("full page missing chrome")
	}
}

// Deck colors come from cards.color_identity (Scryfall color identity), not
// mana_cost: a colorless-cost card with a colored identity still counts, and
// sideboard cards do not contribute.
func TestDeckColorsUseColorIdentity(t *testing.T) {
	h, q := newTestServer(t, "http://127.0.0.1:1")
	srv := httptest.NewServer(h)
	defer srv.Close()
	client := noRedirectClient()
	cookie := registerAndCookie(t, srv, client)

	resp := authedForm(t, srv, client, cookie, http.MethodPost, "/decks", url.Values{
		"name": {"Colors"}, "format": {"40-card"},
	})
	deckPath := resp.Header.Get("Location")
	resp.Body.Close()

	// Colorless mana cost but red identity (e.g. a Devoid card): old
	// mana_cost LIKE logic missed this entirely.
	seedCardFull(t, q, "card-drac", "Dracozoll", "{2}", "R")
	// Green card, ordinary cost.
	seedCardFull(t, q, "card-golgari", "Golgari Thug", "{1}{G}", "G")
	// Blue card in the sideboard only: must not contribute to identity.
	seedCardFull(t, q, "card-island", "Island", "{U}", "U")

	for _, c := range []struct{ id, zone string }{
		{"card-drac", "main"}, {"card-golgari", "main"}, {"card-island", "sideboard"},
	} {
		resp = authedForm(t, srv, client, cookie, http.MethodPost, deckPath+"/cards", url.Values{
			"card_id": {c.id}, "zone": {c.zone}, "quantity": {"1"},
		})
		resp.Body.Close()
	}

	user, err := q.GetUserByEmail(context.Background(), "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := q.ListDeckColors(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListDeckColors rows = %d, want 1", len(rows))
	}
	// WUBRG order → R before G; no U (sideboard) and no W/B.
	if rows[0].Colors != "RG" {
		t.Fatalf("deck colors = %q, want %q", rows[0].Colors, "RG")
	}

	// The decks page renders the identity as mana symbols.
	resp = authedGet(t, srv, client, cookie, "/decks")
	b := body(t, resp)
	for _, sym := range []string{"ms-r", "ms-g"} {
		if !strings.Contains(b, "mana.svg#"+sym) {
			t.Fatalf("decks page missing mana symbol %s", sym)
		}
	}
	for _, sym := range []string{"ms-u", "ms-w", "ms-b"} {
		if strings.Contains(b, "mana.svg#"+sym) {
			t.Fatalf("decks page shows unexpected mana symbol %s", sym)
		}
	}
}

// FR-DECK-07: search results carry an "in <deck>"/"in N decks" badge for
// printings already in the user's decks — and only the user's decks.
func TestSearchMarksDeckMembership(t *testing.T) {
	h, q := newTestServer(t, "http://127.0.0.1:1")
	srv := httptest.NewServer(h)
	defer srv.Close()
	client := noRedirectClient()

	// Two users.
	cookieA := registerAndCookie(t, srv, client)
	resp := postForm(t, srv, client, "/register", url.Values{
		"email": {"other@example.com"}, "display_name": {"Other"}, "password": {"tencharacters!"},
	})
	var cookieB *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "cockadeck_token" {
			cookieB = c
		}
	}
	resp.Body.Close()

	// Three cached printings, each with a unique set+number for lookup.
	seed := func(id, name, num string) {
		t.Helper()
		_, err := q.UpsertCard(context.Background(), sqlcgen.UpsertCardParams{
			ID: id, OracleID: "o-" + id, Name: name, SetCode: "tst",
			SetName: "Test Set", CollectorNumber: num,
			Rarity: "common", Cmc: sql.NullFloat64{Float64: 1, Valid: true},
			ImageUri: "https://img/x.jpg", Raw: "{}",
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	seed("card-a", "Alpha Card", "1")
	seed("card-b", "Beta Card", "2")
	seed("card-c", "Gamma Card", "3")

	newDeck := func(cookie *http.Cookie, name string) string {
		t.Helper()
		resp := authedForm(t, srv, client, cookie, http.MethodPost, "/decks", url.Values{
			"name": {name}, "format": {"40-card"},
		})
		defer resp.Body.Close()
		return resp.Header.Get("Location")
	}
	addCard := func(cookie *http.Cookie, deckPath, cardID string) {
		t.Helper()
		resp := authedForm(t, srv, client, cookie, http.MethodPost, deckPath+"/cards", url.Values{
			"card_id": {cardID}, "zone": {"main"}, "quantity": {"1"},
		})
		resp.Body.Close()
	}

	// User A: card-a in two decks, card-b in one.
	deckA1 := newDeck(cookieA, "Alpha One")
	deckA2 := newDeck(cookieA, "Alpha Two")
	addCard(cookieA, deckA1, "card-a")
	addCard(cookieA, deckA2, "card-a")
	addCard(cookieA, deckA1, "card-b")
	// User B: only card-c.
	deckB := newDeck(cookieB, "Beta Deck")
	addCard(cookieB, deckB, "card-c")

	search := func(cookie *http.Cookie, num string) string {
		t.Helper()
		resp := authedGet(t, srv, client, cookie, "/search?set=tst&num="+num)
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return string(b)
	}

	// A sees aggregate and single-deck badges, none for card-c (B's deck only).
	if b := search(cookieA, "1"); !strings.Contains(b, "in 2 decks") {
		t.Fatalf("card-a badge missing: %.300s", b)
	}
	if b := search(cookieA, "2"); !strings.Contains(b, "in Alpha One") {
		t.Fatalf("card-b badge missing: %.300s", b)
	}
	if b := search(cookieA, "3"); strings.Contains(b, "badge-in-deck") {
		t.Fatal("card-c shows a badge for user A, want none (it is only in B's deck)")
	}
	// B sees only its own membership.
	if b := search(cookieB, "1"); strings.Contains(b, "badge-in-deck") {
		t.Fatal("card-a shows a badge for user B, want none (A's decks don't count)")
	}
	if b := search(cookieB, "3"); !strings.Contains(b, "in Beta Deck") {
		t.Fatalf("card-c badge missing for B: %.300s", b)
	}
}

func TestDeckOwnership(t *testing.T) {
	h, _ := newTestServer(t, "http://127.0.0.1:1")
	srv := httptest.NewServer(h)
	defer srv.Close()
	client := noRedirectClient()

	owner := registerAndCookie(t, srv, client)
	resp := authedForm(t, srv, client, owner, http.MethodPost, "/decks", url.Values{
		"name": {"Mine"}, "format": {"40-card"},
	})
	deckPath := resp.Header.Get("Location")
	resp.Body.Close()

	// Second user cannot see it.
	resp = postForm(t, srv, client, "/register", url.Values{
		"email": {"other@example.com"}, "display_name": {"Other"}, "password": {"tencharacters!"},
	})
	var other *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "cockadeck_token" {
			other = c
		}
	}
	resp.Body.Close()

	resp = authedGet(t, srv, client, other, deckPath)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("foreign deck: got %d, want 404", resp.StatusCode)
	}
}
