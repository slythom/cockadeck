// Package catalog holds the domain services: card lookup with Scryfall
// cache-fill, and (later) deck and collection business rules.
package catalog

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	"cockadeck/internal/scryfall"
	"cockadeck/internal/store/sqlcgen"
)

type Service struct {
	q  sqlcgen.Querier
	sf *scryfall.Client
}

func New(q sqlcgen.Querier, sf *scryfall.Client) *Service {
	return &Service{q: q, sf: sf}
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
