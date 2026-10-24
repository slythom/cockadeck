// Package catalog holds the domain services: card lookup with Scryfall
// cache-fill, and deck/collection business rules.
package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"cockadeck/internal/scryfall"
	"cockadeck/internal/store/sqlcgen"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation error")
)

type Service struct {
	q  sqlcgen.Querier
	db *sql.DB
	sf *scryfall.Client
}

func New(q sqlcgen.Querier, db *sql.DB, sf *scryfall.Client) *Service {
	return &Service{q: q, db: db, sf: sf}
}

// CardBySetNumber resolves a printing by set code + collector number,
// checking the local cards cache before calling Scryfall.
func (s *Service) CardBySetNumber(ctx context.Context, set, number string) (sqlcgen.Card, error) {
	set = strings.ToLower(strings.TrimSpace(set))
	number = strings.TrimSpace(number)

	c, err := s.q.GetCardBySetNumber(ctx, sqlcgen.GetCardBySetNumberParams{
		SetCode:         set,
		CollectorNumber: number,
	})
	if err == nil {
		slog.Debug("card lookup", "set", set, "number", number, "cache_hit", true)
		return c, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return sqlcgen.Card{}, err
	}

	sc, err := s.sf.CardBySetNumber(ctx, set, number)
	if err != nil {
		return sqlcgen.Card{}, err
	}
	slog.Debug("card lookup", "set", set, "number", number, "cache_hit", false)
	return s.cacheCard(ctx, sc)
}

// CardsByName runs a Scryfall name search and fills the cards cache with
// every returned printing, so results can be added to decks/collections
// without another API call. Returns the cached rows and Scryfall's total.
func (s *Service) CardsByName(ctx context.Context, name, set string) ([]sqlcgen.Card, int, error) {
	name = strings.TrimSpace(name)
	set = strings.ToLower(strings.TrimSpace(set))

	res, err := s.sf.SearchByName(ctx, name, set)
	if err != nil {
		return nil, 0, err
	}
	cards := make([]sqlcgen.Card, 0, len(res.Cards))
	for _, sc := range res.Cards {
		c, err := s.cacheCard(ctx, sc)
		if err != nil {
			return nil, 0, err
		}
		cards = append(cards, c)
	}
	return cards, res.TotalCards, nil
}

// cacheCard upserts a Scryfall payload into the cards table.
func (s *Service) cacheCard(ctx context.Context, sc *scryfall.Card) (sqlcgen.Card, error) {
	return s.q.UpsertCard(ctx, sqlcgen.UpsertCardParams{
		ID:              sc.ID,
		OracleID:        sc.OracleID,
		Name:            sc.Name,
		SetCode:         sc.SetCode,
		SetName:         sc.SetName,
		CollectorNumber: sc.CollectorNumber,
		Rarity:          sc.Rarity,
		ManaCost:        sql.NullString{String: sc.ManaCost, Valid: sc.ManaCost != ""},
		Cmc:             sql.NullFloat64{Float64: sc.CMC, Valid: true},
		TypeLine:        sql.NullString{String: sc.TypeLine, Valid: sc.TypeLine != ""},
		ImageUri:        sc.ImageURIs.Normal,
		Raw:             string(sc.Raw),
	})
}

// --- Decks -----------------------------------------------------------------

var DeckFormats = []string{"40-card", "commander"}

func validFormat(f string) bool {
	return f == "40-card" || f == "commander"
}

func (s *Service) ListDecks(ctx context.Context, userID int64) ([]sqlcgen.ListDecksRow, error) {
	return s.q.ListDecks(ctx, userID)
}

// GetDeck fetches a deck scoped to its owner; ErrNotFound if absent or foreign.
func (s *Service) GetDeck(ctx context.Context, deckID, userID int64) (sqlcgen.Deck, error) {
	d, err := s.q.GetDeck(ctx, sqlcgen.GetDeckParams{ID: deckID, UserID: userID})
	if errors.Is(err, sql.ErrNoRows) {
		return sqlcgen.Deck{}, ErrNotFound
	}
	return d, err
}

func (s *Service) CreateDeck(ctx context.Context, userID int64, name, description, format string) (sqlcgen.Deck, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return sqlcgen.Deck{}, fmt.Errorf("%w: deck name is required", ErrValidation)
	}
	if !validFormat(format) {
		return sqlcgen.Deck{}, fmt.Errorf("%w: invalid format %q", ErrValidation, format)
	}
	return s.q.CreateDeck(ctx, sqlcgen.CreateDeckParams{
		UserID:      userID,
		Name:        name,
		Description: strings.TrimSpace(description),
		Format:      format,
	})
}

func (s *Service) UpdateDeck(ctx context.Context, deckID, userID int64, name, description, format string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: deck name is required", ErrValidation)
	}
	if !validFormat(format) {
		return fmt.Errorf("%w: invalid format %q", ErrValidation, format)
	}
	if _, err := s.GetDeck(ctx, deckID, userID); err != nil {
		return err
	}
	return s.q.UpdateDeck(ctx, sqlcgen.UpdateDeckParams{
		ID:          deckID,
		UserID:      userID,
		Name:        name,
		Description: strings.TrimSpace(description),
		Format:      format,
	})
}

func (s *Service) DeleteDeck(ctx context.Context, deckID, userID int64) error {
	if _, err := s.GetDeck(ctx, deckID, userID); err != nil {
		return err
	}
	return s.q.DeleteDeck(ctx, sqlcgen.DeleteDeckParams{ID: deckID, UserID: userID})
}

// --- Deck cards --------------------------------------------------------------

// CardFilters are the combinable deck-view filters of FR-DECK-06.
// Empty/nil fields are ignored.
type CardFilters struct {
	Set    string
	Rarity string
	CMC    *float64
}

func (f CardFilters) Params(deckID int64) sqlcgen.ListDeckCardsParams {
	p := sqlcgen.ListDeckCardsParams{DeckID: deckID}
	if f.Set != "" {
		p.SetFilter = f.Set
	}
	if f.Rarity != "" {
		p.Rarity = f.Rarity
	}
	if f.CMC != nil {
		p.Cmc = *f.CMC
	}
	return p
}

func (s *Service) DeckCards(ctx context.Context, deckID int64, f CardFilters) ([]sqlcgen.ListDeckCardsRow, error) {
	return s.q.ListDeckCards(ctx, f.Params(deckID))
}

var validZones = map[string]bool{"main": true, "sideboard": true, "commander": true}

// checkZone enforces the zone set and the commander-zone format rule (FR-DECK-03).
func checkZone(deck sqlcgen.Deck, zone string) error {
	if !validZones[zone] {
		return fmt.Errorf("%w: invalid zone %q", ErrValidation, zone)
	}
	if zone == "commander" && deck.Format != "commander" {
		return fmt.Errorf("%w: commander zone only on commander-format decks", ErrValidation)
	}
	return nil
}

func (s *Service) AddDeckCard(ctx context.Context, deckID, userID int64, cardID, zone string, quantity int64) error {
	if quantity < 1 {
		return fmt.Errorf("%w: quantity must be at least 1", ErrValidation)
	}
	deck, err := s.GetDeck(ctx, deckID, userID)
	if err != nil {
		return err
	}
	if err := checkZone(deck, zone); err != nil {
		return err
	}
	if _, err := s.q.GetCardByID(ctx, cardID); err != nil {
		return fmt.Errorf("%w: unknown card", ErrValidation)
	}
	return s.q.UpsertDeckCard(ctx, sqlcgen.UpsertDeckCardParams{
		DeckID:   deckID,
		CardID:   cardID,
		Zone:     zone,
		Quantity: quantity,
	})
}

// UpdateDeckCard changes a deck row's quantity and/or moves it to a new zone.
// quantity <= 0 removes the row entirely.
func (s *Service) UpdateDeckCard(ctx context.Context, deckID, userID int64, cardID, zone string, quantity *int64, newZone string) error {
	deck, err := s.GetDeck(ctx, deckID, userID)
	if err != nil {
		return err
	}
	if newZone == "" {
		newZone = zone
	}
	if err := checkZone(deck, newZone); err != nil {
		return err
	}
	if quantity != nil && *quantity <= 0 {
		return s.q.DeleteDeckCard(ctx, sqlcgen.DeleteDeckCardParams{
			DeckID: deckID, CardID: cardID, Zone: zone,
		})
	}
	if newZone == zone {
		if quantity == nil {
			return fmt.Errorf("%w: nothing to update", ErrValidation)
		}
		return s.q.SetDeckCardQuantity(ctx, sqlcgen.SetDeckCardQuantityParams{
			DeckID: deckID, CardID: cardID, Zone: zone, Quantity: *quantity,
		})
	}
	return s.moveDeckCard(ctx, deckID, cardID, zone, newZone, quantity)
}

// moveDeckCard merges the source row into the target zone inside a transaction:
// the PK includes zone, so a move is an upsert-into-target + delete-source.
func (s *Service) moveDeckCard(ctx context.Context, deckID int64, cardID, fromZone, toZone string, quantity *int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := sqlcgen.New(tx)

	row, err := qtx.GetDeckCard(ctx, sqlcgen.GetDeckCardParams{
		DeckID: deckID, CardID: cardID, Zone: fromZone,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	qty := row.Quantity
	if quantity != nil {
		qty = *quantity
	}
	if err := qtx.UpsertDeckCard(ctx, sqlcgen.UpsertDeckCardParams{
		DeckID: deckID, CardID: cardID, Zone: toZone, Quantity: qty,
	}); err != nil {
		return err
	}
	if err := qtx.DeleteDeckCard(ctx, sqlcgen.DeleteDeckCardParams{
		DeckID: deckID, CardID: cardID, Zone: fromZone,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) RemoveDeckCard(ctx context.Context, deckID, userID int64, cardID, zone string) error {
	if _, err := s.GetDeck(ctx, deckID, userID); err != nil {
		return err
	}
	return s.q.DeleteDeckCard(ctx, sqlcgen.DeleteDeckCardParams{
		DeckID: deckID, CardID: cardID, Zone: zone,
	})
}
