-- +goose Up
CREATE TABLE users (
    id            INTEGER PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE COLLATE NOCASE,
    display_name  TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE cards (
    id               TEXT PRIMARY KEY,
    oracle_id        TEXT NOT NULL,
    name             TEXT NOT NULL,
    set_code         TEXT NOT NULL,
    set_name         TEXT NOT NULL,
    collector_number TEXT NOT NULL,
    rarity           TEXT NOT NULL,
    mana_cost        TEXT,
    cmc              REAL,
    type_line        TEXT,
    image_uri        TEXT NOT NULL,
    raw              TEXT NOT NULL,
    fetched_at       TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE decks (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    format      TEXT NOT NULL CHECK (format IN ('40-card', 'commander')),
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE deck_cards (
    deck_id  INTEGER NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    card_id  TEXT NOT NULL REFERENCES cards(id),
    zone     TEXT NOT NULL DEFAULT 'main' CHECK (zone IN ('main', 'sideboard', 'commander')),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (deck_id, card_id, zone)
);

CREATE TABLE collections (
    id        INTEGER PRIMARY KEY,
    user_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name      TEXT NOT NULL DEFAULT 'Default',
    is_shared INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE collection_cards (
    collection_id INTEGER NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    card_id       TEXT NOT NULL REFERENCES cards(id),
    finish        TEXT NOT NULL DEFAULT 'nonfoil' CHECK (finish IN ('nonfoil', 'foil')),
    quantity      INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (collection_id, card_id, finish)
);

CREATE INDEX idx_cards_set_number ON cards (set_code, collector_number);
CREATE INDEX idx_cards_name ON cards (name);
CREATE INDEX idx_deck_cards_card ON deck_cards (card_id);
CREATE INDEX idx_collection_cards_card ON collection_cards (card_id);
CREATE INDEX idx_collections_user ON collections (user_id);
CREATE INDEX idx_decks_user ON decks (user_id);

-- +goose Down
DROP TABLE collection_cards;
DROP TABLE collections;
DROP TABLE deck_cards;
DROP TABLE decks;
DROP TABLE cards;
DROP TABLE users;
