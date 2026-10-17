// Package scryfall is a rate-limited client for the Scryfall API.
package scryfall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

var (
	ErrNotFound     = errors.New("card not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnavailable  = errors.New("scryfall unavailable")
)

type Client struct {
	http    *http.Client
	baseURL string
	limiter *rate.Limiter
}

func NewClient(baseURL string) *Client {
	return &Client{
		http:    &http.Client{Timeout: 10 * time.Second},
		baseURL: strings.TrimRight(baseURL, "/"),
		limiter: rate.NewLimiter(rate.Every(75*time.Millisecond), 1),
	}
}

type imageURIs struct {
	Normal string `json:"normal"`
}

type cardFace struct {
	ManaCost  string    `json:"mana_cost"`
	TypeLine  string    `json:"type_line"`
	ImageURIs imageURIs `json:"image_uris"`
}

type Card struct {
	ID              string     `json:"id"`
	OracleID        string     `json:"oracle_id"`
	Name            string     `json:"name"`
	SetCode         string     `json:"set"`
	SetName         string     `json:"set_name"`
	CollectorNumber string     `json:"collector_number"`
	Rarity          string     `json:"rarity"`
	ManaCost        string     `json:"mana_cost"`
	CMC             float64    `json:"cmc"`
	TypeLine        string     `json:"type_line"`
	ImageURIs       imageURIs  `json:"image_uris"`
	CardFaces       []cardFace `json:"card_faces"`

	Raw []byte `json:"-"`
}

// fillFromFaces backfills fields that live on card_faces for double-faced cards.
func (c *Card) fillFromFaces() {
	if len(c.CardFaces) == 0 {
		return
	}
	f := c.CardFaces[0]
	if c.ManaCost == "" {
		c.ManaCost = f.ManaCost
	}
	if c.TypeLine == "" {
		c.TypeLine = f.TypeLine
	}
	if c.ImageURIs.Normal == "" {
		c.ImageURIs.Normal = f.ImageURIs.Normal
	}
}

func (c *Client) CardBySetNumber(ctx context.Context, set, number string) (*Card, error) {
	path := "/cards/" + url.PathEscape(strings.ToLower(set)) + "/" + url.PathEscape(number)
	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	var card Card
	if err := json.Unmarshal(body, &card); err != nil {
		return nil, fmt.Errorf("%w: malformed response: %v", ErrUnavailable, err)
	}
	card.Raw = body
	card.fillFromFaces()
	return &card, nil
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	req.Header.Set("User-Agent", "Cockadeck/0.1")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	switch {
	case resp.StatusCode == http.StatusOK:
		return body, nil
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, scryfallErrorDetail(body))
	default:
		return nil, fmt.Errorf("%w: status %d", ErrUnavailable, resp.StatusCode)
	}
}

func scryfallErrorDetail(body []byte) string {
	var e struct {
		Details string `json:"details"`
	}
	if json.Unmarshal(body, &e) == nil {
		return e.Details
	}
	return ""
}
