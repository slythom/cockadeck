-- name: GetCardBySetNumber :one
SELECT id, oracle_id, name, set_code, set_name, collector_number, rarity,
       mana_cost, cmc, type_line, image_uri, raw, fetched_at
FROM cards
WHERE set_code = sqlc.arg(set_code) AND collector_number = sqlc.arg(collector_number);

-- name: GetCardByID :one
SELECT id, oracle_id, name, set_code, set_name, collector_number, rarity,
       mana_cost, cmc, type_line, image_uri, raw, fetched_at
FROM cards
WHERE id = ?;

-- name: UpsertCard :one
INSERT INTO cards (id, oracle_id, name, set_code, set_name, collector_number,
                   rarity, mana_cost, cmc, type_line, image_uri, raw)
VALUES (sqlc.arg(id), sqlc.arg(oracle_id), sqlc.arg(name), sqlc.arg(set_code),
        sqlc.arg(set_name), sqlc.arg(collector_number), sqlc.arg(rarity),
        sqlc.arg(mana_cost), sqlc.arg(cmc), sqlc.arg(type_line),
        sqlc.arg(image_uri), sqlc.arg(raw))
ON CONFLICT (id) DO UPDATE SET
    oracle_id        = excluded.oracle_id,
    name             = excluded.name,
    set_code         = excluded.set_code,
    set_name         = excluded.set_name,
    collector_number = excluded.collector_number,
    rarity           = excluded.rarity,
    mana_cost        = excluded.mana_cost,
    cmc              = excluded.cmc,
    type_line        = excluded.type_line,
    image_uri        = excluded.image_uri,
    raw              = excluded.raw,
    fetched_at       = datetime('now')
RETURNING id, oracle_id, name, set_code, set_name, collector_number, rarity,
          mana_cost, cmc, type_line, image_uri, raw, fetched_at;
