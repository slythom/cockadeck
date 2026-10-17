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
	_, err := q.UpsertCard(context.Background(), sqlcgen.UpsertCardParams{
		ID: id, OracleID: "o-" + id, Name: name, SetCode: "otj",
		SetName: "Outlaws of Thunder Junction", CollectorNumber: "1",
		Rarity: "common", Cmc: sql.NullFloat64{Float64: 2, Valid: true},
		ImageUri: "https://img/x.jpg", Raw: "{}",
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

	// Remove.
	resp = authedForm(t, srv, client, cookie, http.MethodDelete, deckPath+"/cards/card-1?zone=commander", nil)
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
