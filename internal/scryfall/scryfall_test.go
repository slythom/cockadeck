package scryfall

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fixture(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient(srv.URL)
}

func TestCardBySetNumber(t *testing.T) {
	var gotPath, gotUA string
	c := fixture(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{
			"id": "id-1", "oracle_id": "o-1", "name": "Sol Ring",
			"set": "mkm", "set_name": "Murders at Karlov Manor",
			"collector_number": "361\u2605", "rarity": "mythic",
			"image_uris": {"normal": "https://img/x.jpg"}
		}`)
	})

	card, err := c.CardBySetNumber(context.Background(), "MKM", "361★")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/cards/mkm/361★" {
		t.Fatalf("path = %q, special collector number not escaped correctly", gotPath)
	}
	if gotUA != "Cockadeck/0.1" {
		t.Fatalf("User-Agent = %q", gotUA)
	}
	if card.ID != "id-1" || card.CollectorNumber != "361★" || len(card.Raw) == 0 {
		t.Fatalf("card = %+v", card)
	}
}

func TestCardFacesFallback(t *testing.T) {
	c := fixture(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{
			"id": "dfc-1", "oracle_id": "o-2", "name": "Front // Back",
			"set": "dsk", "set_name": "Duskmourn", "collector_number": "1",
			"rarity": "rare", "cmc": 3.0,
			"card_faces": [{
				"mana_cost": "{2}{R}", "type_line": "Enchantment — Room",
				"image_uris": {"normal": "https://img/front.jpg"}
			}, {
				"mana_cost": "{1}{U}", "type_line": "Enchantment — Room",
				"image_uris": {"normal": "https://img/back.jpg"}
			}]
		}`)
	})

	card, err := c.CardBySetNumber(context.Background(), "dsk", "1")
	if err != nil {
		t.Fatal(err)
	}
	if card.ManaCost != "{2}{R}" || card.TypeLine != "Enchantment — Room" || card.ImageURIs.Normal != "https://img/front.jpg" {
		t.Fatalf("face fallback failed: %+v", card)
	}
}

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   error
	}{
		{http.StatusNotFound, `{"object":"error","details":"not found"}`, ErrNotFound},
		{http.StatusBadRequest, `{"object":"error","details":"bad set code"}`, ErrInvalidInput},
		{http.StatusInternalServerError, `{}`, ErrUnavailable},
	}
	for _, tc := range cases {
		c := fixture(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(tc.status)
			io.WriteString(w, tc.body)
		})
		_, err := c.CardBySetNumber(context.Background(), "x", "1")
		if !errors.Is(err, tc.want) {
			t.Fatalf("status %d: got %v, want %v", tc.status, err, tc.want)
		}
	}
}
