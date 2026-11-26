-- +goose Up
ALTER TABLE cards ADD COLUMN color_identity TEXT NOT NULL DEFAULT '';

-- Backfill from the cached Scryfall payload: color_identity is a JSON array
-- of single letters ("W","U","B","R","G"); store them concatenated ("WUB").
UPDATE cards
SET color_identity = COALESCE(
    (SELECT group_concat(j.value, '')
     FROM json_each(cards.raw, '$.color_identity') AS j),
    '');

-- +goose Down
ALTER TABLE cards DROP COLUMN color_identity;
