-- name: ListDecks :many
SELECT d.id, d.user_id, d.name, d.description, d.format, d.created_at, d.updated_at,
       COALESCE(SUM(dc.quantity), 0) AS card_count
FROM decks d
LEFT JOIN deck_cards dc ON dc.deck_id = d.id
WHERE d.user_id = ?
GROUP BY d.id
ORDER BY d.updated_at DESC, d.id DESC;

-- name: GetDeck :one
SELECT id, user_id, name, description, format, created_at, updated_at
FROM decks
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: CreateDeck :one
INSERT INTO decks (user_id, name, description, format)
VALUES (sqlc.arg(user_id), sqlc.arg(name), sqlc.arg(description), sqlc.arg(format))
RETURNING id, user_id, name, description, format, created_at, updated_at;

-- name: UpdateDeck :exec
UPDATE decks
SET name = sqlc.arg(name), description = sqlc.arg(description),
    format = sqlc.arg(format), updated_at = datetime('now')
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: DeleteDeck :exec
DELETE FROM decks
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: GetDeckCard :one
SELECT deck_id, card_id, zone, quantity
FROM deck_cards
WHERE deck_id = sqlc.arg(deck_id) AND card_id = sqlc.arg(card_id) AND zone = sqlc.arg(zone);

-- name: ListDeckCards :many
SELECT dc.deck_id, dc.card_id, dc.zone, dc.quantity,
       c.name, c.set_code, c.set_name, c.collector_number, c.rarity,
       c.mana_cost, c.cmc, c.type_line, c.image_uri
FROM deck_cards dc
JOIN cards c ON c.id = dc.card_id
WHERE dc.deck_id = sqlc.arg(deck_id)
  AND (sqlc.narg(set_filter) IS NULL
       OR c.set_code = sqlc.narg(set_filter)
       OR lower(c.set_name) = lower(sqlc.narg(set_filter)))
  AND (sqlc.narg(rarity) IS NULL OR c.rarity = sqlc.narg(rarity))
  AND (sqlc.narg(cmc) IS NULL OR c.cmc = sqlc.narg(cmc))
ORDER BY c.cmc, c.name;

-- name: UpsertDeckCard :exec
INSERT INTO deck_cards (deck_id, card_id, zone, quantity)
VALUES (sqlc.arg(deck_id), sqlc.arg(card_id), sqlc.arg(zone), sqlc.arg(quantity))
ON CONFLICT (deck_id, card_id, zone)
DO UPDATE SET quantity = deck_cards.quantity + excluded.quantity;

-- name: SetDeckCardQuantity :exec
UPDATE deck_cards SET quantity = sqlc.arg(quantity)
WHERE deck_id = sqlc.arg(deck_id) AND card_id = sqlc.arg(card_id) AND zone = sqlc.arg(zone);

-- name: DeleteDeckCard :exec
DELETE FROM deck_cards
WHERE deck_id = sqlc.arg(deck_id) AND card_id = sqlc.arg(card_id) AND zone = sqlc.arg(zone);
