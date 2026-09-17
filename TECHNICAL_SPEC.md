# Cockadeck — Technical Specification

| | |
|---|---|
| Status | Draft v1 |
| Date | 2026-09-17 |
| Companion document | `SPEC.md` (feature spec — source of truth for *what*) |
| Scope | *How* the features in SPEC.md are built: stack, architecture, schema, routes, integrations |

---

## 1. Architecture Overview

Cockadeck is a **single self-contained Go binary**: a server-rendered web application where

- **Go `net/http`** (stdlib, no web framework) serves both full pages and HTML fragments.
- **Templ** compiles type-safe templates to Go.
- **HTMX** (vendored static asset) drives partial-page updates — fragments are returned by the same endpoints, selected via the `HX-Request` header.
- **SQLite** is the only datastore, accessed through **sqlc-generated** code. The DB file lives next to the binary on the VPS.
- **Scryfall** is the sole external dependency, called through a rate-limited, caching client.
- Card images are **hotlinked** from Scryfall's CDN — no image storage or proxying.

```
Browser ──HTMX──▶ nginx (TLS) ──▶ cockadeck (Go)
                                   ├─ templ → HTML / fragments
                                   ├─ sqlc → SQLite (WAL)
                                   └─ scryfall client → api.scryfall.com (rate-limited, cached in SQLite)
```

Deliberately **not** used: no SPA framework, no ORM, no router library, no Redis, no separate frontend build step.

### Request layering

```
handlers (internal/http) → domain services → store (sqlc) → SQLite
                       └→ scryfall client (lookup + cache fill)
```

Thin handlers: parse input → call service → render Templ. Services hold business rules (zones, finishes, tradable logic). The sqlc `Queries` type is the repository — no extra repository abstraction until it earns its place. Manual constructor injection in `main` — no DI framework.

---

## 2. Tech Stack

| Layer | Choice | Notes |
|---|---|---|
| Language | Go 1.26 | single static binary |
| HTTP | `net/http` + `ServeMux` | Go 1.22+ pattern routing (`GET /decks/{id}`) |
| Templates | `github.com/a-h/templ` | already in `go.mod` |
| Interactivity | HTMX (vendored `htmx.min.js`) | `hx-get`/`hx-post`/`hx-swap`/`hx-trigger` |
| CSS | custom, shadcn/ui-inspired | one stylesheet; no Tailwind/build step (revisit if it becomes painful — decision TD9) |
| JS | inline snippets only when unavoidable | e.g. `<dialog>` wiring |
| Database | SQLite via `modernc.org/sqlite` | pure-Go driver, no cgo → trivial builds |
| Queries | sqlc (`engine: sqlite`) | generated type-safe Go from `.sql` files |
| Migrations | goose (`pressly/goose/v3`) | versioned `.sql` migrations embedded with `go:embed`, applied at startup (decision TD8) |
| Auth | JWT (`golang-jwt/jwt/v5`) in `httpOnly` cookie | argon2id password hashing (`golang.org/x/crypto`) |
| Scryfall client | stdlib `net/http` + `x/time/rate` | ~75 ms min spacing, local cache |
| Logging | `log/slog` (stdlib) | JSON handler in prod |
| Telemetry | OpenTelemetry → SigNoz | **post-MVP**; structured logs first (§11) |
| Hosting | small VPS, Ubuntu 26.04 | systemd unit |
| TLS | nginx reverse proxy + certbot | (decision TD7) |

### Dependency list (complete for MVP)

```
github.com/a-h/templ
github.com/golang-jwt/jwt/v5
golang.org/x/crypto        # argon2id
golang.org/x/time          # rate limiter
modernc.org/sqlite
github.com/pressly/goose/v3  # migrations (also usable as CLI)
```

Build-time tools (not runtime deps): `templ` CLI, `sqlc` CLI. HTMX is vendored under `static/` — not a Go dependency.

---

## 3. Project Layout

Right-sized for a personal project: `internal/` packages by concern, no hexagonal ceremony.

```
cockadeck/
├── cmd/cockadeck/main.go        # wiring only: config → db → server
├── internal/
│   ├── config/                  # env-based config (12-factor)
│   ├── auth/                    # password hashing, JWT issue/verify, middleware
│   ├── server/                  # router, handlers, middleware, HTMX helpers
│   ├── scryfall/                # API client, rate limiter, types, cache fill
│   ├── catalog/                 # deck/collection/search domain services
│   ├── store/
│   │   ├── db.go                # sql.Open, pragmas, pool config
│   │   └── sqlcgen/             # generated — do not edit
│   └── export/                  # .cod (XML) and CSV writers
├── db/
│   ├── migrations/              # goose .sql files
│   └── queries/                 # sqlc input .sql files
├── views/                       # .templ (layout, pages, components)
├── static/                      # app.css, htmx.min.js (embedded via go:embed)
├── sqlc.yaml
├── Makefile                     # generate, build, test, migrate, dev
└── SPEC.md / TECHNICAL_SPEC.md
```

The existing `router/` and `handlers/` packages fold into `internal/server`; `views/` stays at root.

---

## 4. Configuration

12-factor: env vars only, read once in `internal/config`.

| Var | Default | Purpose |
|---|---|---|
| `COCKADECK_ADDR` | `:8080` | listen address |
| `COCKADECK_DB_PATH` | `./cockadeck.db` | SQLite file |
| `COCKADECK_JWT_SECRET` | *(required)* | HMAC key for token signing |
| `COCKADECK_ENV` | `dev` | `dev`/`prod` — controls Secure cookie, log format |
| `COCKADECK_SCRYFALL_URL` | `https://api.scryfall.com` | override for tests |

No config files, no secrets in the repo. Secret supplied via systemd `EnvironmentFile`.

---

## 5. Database

### 5.1 Engine setup

- `modernc.org/sqlite`, `database/sql`, single `*sql.DB`.
- Pragmas on open: `journal_mode=WAL`, `foreign_keys=ON`, `busy_timeout=5000`, `synchronous=NORMAL`.
- `SetMaxOpenConns(1)` for writes is unnecessary under WAL for this scale, but keep pool small (`SetMaxOpenConns(4)`); revisit only if contention appears.
- Migrations run at startup via goose with embedded FS — one binary, no external migrate step.

### 5.2 sqlc

`sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "sqlite"
    schema: "db/migrations"
    queries: "db/queries"
    gen:
      go:
        package: "sqlcgen"
        out: "internal/store/sqlcgen"
        emit_json_tags: true
        emit_interface: true
```

`emit_interface` gives a `Querier` interface for free → services depend on it, tests can fake it.

### 5.3 Schema (draft — to be reviewed before first migration)

The `cards` table is the **local Scryfall cache**, not a card catalog: rows appear when a card is first fetched, and everything else references it.

```sql
users (
  id            INTEGER PRIMARY KEY,
  email         TEXT NOT NULL UNIQUE COLLATE NOCASE,
  display_name  TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  created_at    TEXT NOT NULL DEFAULT (datetime('now'))
)

cards (                                  -- Scryfall cache; id = Scryfall UUID
  id               TEXT PRIMARY KEY,     -- scryfall id
  oracle_id        TEXT NOT NULL,
  name             TEXT NOT NULL,
  set_code         TEXT NOT NULL,
  set_name         TEXT NOT NULL,
  collector_number TEXT NOT NULL,
  rarity           TEXT NOT NULL,
  mana_cost        TEXT,
  cmc              REAL,
  type_line        TEXT,
  image_uri        TEXT NOT NULL,        -- image_uris.normal
  raw              TEXT NOT NULL,        -- full JSON payload, for future fields
  fetched_at       TEXT NOT NULL DEFAULT (datetime('now'))
)

decks (
  id          INTEGER PRIMARY KEY,
  user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  format      TEXT NOT NULL CHECK (format IN ('40-card','commander')),
  created_at / updated_at
)

deck_cards (
  deck_id  INTEGER NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
  card_id  TEXT NOT NULL REFERENCES cards(id),
  zone     TEXT NOT NULL DEFAULT 'main' CHECK (zone IN ('main','sideboard','commander')),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  PRIMARY KEY (deck_id, card_id, zone)   -- same card in two zones = two rows
)

collections (
  id        INTEGER PRIMARY KEY,
  user_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name      TEXT NOT NULL DEFAULT 'Default',
  is_shared INTEGER NOT NULL DEFAULT 0   -- boolean
)

collection_cards (
  collection_id INTEGER NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
  card_id       TEXT NOT NULL REFERENCES cards(id),
  finish        TEXT NOT NULL DEFAULT 'nonfoil' CHECK (finish IN ('nonfoil','foil')),
  quantity      INTEGER NOT NULL CHECK (quantity > 0),
  PRIMARY KEY (collection_id, card_id, finish)
)
```

Indexes beyond PKs: `cards(set_code, collector_number)` (dedupe/lookup), `cards(name)` (tradable join), `deck_cards(card_id)`, `collection_cards(card_id)`, `collections(user_id)`, `decks(user_id)`.

**Design notes**

- `deck_cards` uniqueness includes `zone` — moving a card between zones is an upsert/delete-merge, not an UPDATE.
- `raw` JSON keeps every Scryfall field available later (stats, legalities) without migrations.
- No `sessions` table — JWT is stateless (see §6 and TD2).
- Import (SPEC §5.8) needs no schema change: it produces normal `deck_cards`/`collection_cards` writes.

### 5.4 Key queries

- **Collection view with filters**: `WHERE` clauses built via sqlc `sqlc.narg()` optional params for set/rarity/finish; mana-cost filtering on `cards.cmc`.
- **Set completion (FR-COLL-05)**: after `scryfall` client fills `cards` with the full set, `LEFT JOIN collection_cards` on `(card_id)` → `owned`/`missing` per printing.
- **Tradable (FR-CMP-03)**, name-level, one-way — computed in SQL:

```sql
SELECT cc.*, c.*,
  NOT EXISTS (
    SELECT 1 FROM collection_cards mine
    JOIN collections mc ON mc.id = mine.collection_id AND mc.user_id = :viewer_id
    JOIN cards cn ON cn.id = mine.card_id
    WHERE cn.name = c.name AND mine.quantity > 0
  ) AS tradable
FROM collection_cards cc
JOIN cards c ON c.id = cc.card_id
JOIN collections col ON col.id = cc.collection_id
WHERE col.user_id = :owner_id AND col.is_shared = 1;
```

The `is_shared` predicate is in the query itself — a private collection returns zero rows, so the check can't be bypassed by handler bugs (NFR-04).

---

## 6. Authentication

| Aspect | Design |
|---|---|
| Registration | email + password → argon2id hash → `users` row; email `UNIQUE COLLATE NOCASE` (FR-AUTH-05) |
| Password rules | min length 10; validated server-side (FR-AUTH-05) |
| Token | JWT, HS256, claims `sub` (user id), `exp`, `iat`; 30-day expiry |
| Transport | cookie `cockadeck_token`: `HttpOnly`, `Secure` (prod), `SameSite=Lax`, `Path=/` |
| Middleware | `RequireAuth` wraps everything except `/login`, `/register`, `/static/*`; parses + verifies JWT, puts user id in `context.Context`; expired/missing → redirect `/login` (full page) or `401` + `HX-Redirect` (HTMX request) |
| Logout | expire the cookie |
| Authorization | every query is scoped by `user_id` from context — never from the request; shared-collection reads are the only cross-user path and go through the §5.4 query |

**TD2 — JWT vs server-side sessions.** JWT is the stated choice and is fine at this scale, but note the honest trade-off: stateless tokens can't be revoked before expiry, and a sessions table in SQLite would be ~20 lines with logout-everywhere for free. JWT pays off only if SSO/Keycloak arrives later — at which point it becomes OIDC anyway. Flagged as an open decision; spec proceeds with JWT.

**CSRF**: `SameSite=Lax` plus an `Origin`/`Referer` check middleware on all `POST`/`PUT`/`PATCH`/`DELETE` — sufficient for a cookie-auth HTMX app without token plumbing.

---

## 7. Scryfall Integration (`internal/scryfall`)

### 7.1 Client

- One `*http.Client`, `10s` timeout, `User-Agent: Cockadeck/0.1` per Scryfall etiquette.
- Rate limiter: `x/time/rate.NewLimiter(rate.Every(75*time.Millisecond), 1)` — ~13 req/s ceiling is above guidance but the burst of 1 plus serial calls keep real traffic well under 10/s (NFR-05). Await limiter before every request.
- Errors mapped to domain results: `404` → *card not found*, `4xx` → *invalid input*, `5xx`/timeout → *Scryfall unavailable* (FR-SEARCH-04).
- All fetches take `context.Context` — request-scoped cancellation.

### 7.2 Endpoints used

| Feature | Endpoint |
|---|---|
| Set + collector lookup (FR-SEARCH-01) | `GET /cards/{set}/{number}` (number URL-escaped — handles `361★`, `5a`) |
| Name search (FR-SEARCH-06) | `GET /cards/search?q={name}` |
| Other printings (FR-SEARCH-07) | `GET /cards/search?q=!"{name}" unique:prints` or `prints_search_uri` |
| Set completion (FR-COLL-05) | `GET /cards/search?q=set:{code} unique:prints&order=collector_number`, paginated via `has_more`/`next_page` |

### 7.3 Caching

- Every card payload returned is **upserted into `cards`** (cache fill), so the app can later serve deck/collection views without re-hitting Scryfall.
- Lookups check the `cards` table first (`set_code` + `collector_number`); on miss → API → insert → return.
- Printings are immutable in practice: no TTL in MVP. A `fetched_at`-based refresh can come later if needed.
- Set-completion fetch is the only bulk operation: one paginated crawl per set view, cached afterward. Acceptable at friends scale.
- Images: templates emit `<img src={card.ImageURI}>` pointing at `images.scryfall.com` — hotlinked, `loading="lazy"`.

---

## 8. HTTP Surface (`internal/server`)

Route table. Fragments = same endpoint returning a partial when `HX-Request: true`.

### Pages (full HTML)

| Method | Path | Purpose |
|---|---|---|
| GET | `/login`, `/register` | auth pages (only unauthenticated routes) |
| POST | `/login`, `/register`, `/logout` | auth actions |
| GET | `/` | home → redirect `/decks` |
| GET | `/search` | search page |
| GET | `/decks` | deck list |
| GET | `/decks/{id}` | deck detail (zone-grouped) |
| GET | `/collection` | own collection |
| GET | `/collections` | users with shared collections (FR-CMP-02) |
| GET | `/collections/{userID}` | read-only view of a shared collection (FR-CMP-03/05) |

### Actions / fragments

| Method | Path | Purpose |
|---|---|---|
| GET | `/search/results?set=&num=` or `?name=` | card result(s) fragment; `hx-trigger="keyup changed delay:300ms"` for name search |
| POST | `/decks` | create deck → `HX-Redirect` |
| POST | `/decks/{id}/cards` | add card (`card_id`, `zone`, `quantity`) → updated deck fragment or toast |
| PATCH | `/decks/{id}/cards/{cardID}` | quantity/zone change (zone in query/body) → row fragment |
| DELETE | `/decks/{id}/cards/{cardID}` | remove → empty/200 swap |
| POST/DELETE | `/decks/{id}` (edit/rename), DELETE `/decks/{id}` | deck CRUD |
| POST | `/collection/cards` | add (`card_id`, `finish`, `quantity`) |
| PATCH/DELETE | `/collection/cards/{cardID}` | quantity/remove (finish as key param) |
| POST | `/collection/sharing` | toggle `is_shared` (FR-CMP-01) |
| GET | `/collection?set=otj` | same page, fragment refresh on filter change; `set` param triggers set-completion mode |

Filters (FR-DECK-06, FR-COLL-04, FR-CMP-04) are **query params** — shareable URLs, and HTMX just re-requests the page body fragment.

### Downloads

| Method | Path | Purpose |
|---|---|---|
| GET | `/decks/{id}/export.cod` | Cockatrice XML (§9.1) |
| GET | `/collection/export.csv` | collection CSV (§9.2) |

### Middleware chain (outer → inner)

`Recover` → `RequestLog (slog)` → `SecureHeaders` → `RequireAuth` (skipped for public paths) → `CSRFOriginCheck` → handler.

---

## 9. Export Formats (`internal/export`)

### 9.1 Cockatrice `.cod`

XML via `encoding/xml`. Zones `main`/`side`; in-app `commander` zone **merges into `main`** (D12). Deck name → `<deckname>`, description → `<comments>` (FR-EXP-02).

```xml
<?xml version="1.0" encoding="UTF-8"?>
<cockatrice_deck version="1">
  <deckname>My Deck</deckname>
  <comments>...</comments>
  <zone name="main">
    <card number="4" name="Lightning Bolt"/>
  </zone>
  <zone name="side">...</zone>
</cockatrice_deck>
```

Headers: `Content-Type: application/octet-stream`, `Content-Disposition: attachment; filename="{slug}.cod"` (FR-EXP-04).

### 9.2 Collection CSV

`encoding/csv`, one row per `(printing, finish)` (FR-EXP-03):

```
name,set_code,set_name,collector_number,rarity,finish,quantity
Lightning Bolt,m11,Core Set 2011,146,rare,nonfoil,2
```

Filename `collection-YYYY-MM-DD.csv`.

---

## 10. Frontend

### 10.1 Templ structure

```
views/
├── layout.templ        # <html> shell: css, htmx, nav, flash/toast region
├── auth.templ          # login/register
├── search.templ        # page + Results() fragment component
├── decks.templ         # list + detail + DeckCardRow() fragment
├── collection.templ    # own collection + CollectionRow() fragment
└── compare.templ       # shared-collection browser + tradable view
```

Page components render full layout; fragment components render inner blocks. Handler picks by `HX-Request` header (small `IsHTMX(r)` helper).

### 10.2 HTMX patterns

- **Search-as-you-type**: `hx-get="/search/results" hx-trigger="keyup changed delay:300ms" hx-target="#results"`.
- **Add to deck/collection**: one form per result with two submit paths (`hx-post` with `hx-vals`), or a small popover (`<details>`/`<dialog>`) listing decks + finish/qty — deck and collection targets simultaneously per §7.5 of SPEC.
- **Quantity +/-**: `hx-patch` on buttons, `hx-swap="outerHTML"` on the row — server returns the re-rendered row (single source of truth for quantity/totals).
- **Filters**: a form with `hx-get`, `hx-trigger="change"`, targeting the list container; `hx-push-url="true"` keeps URLs shareable.
- **Sharing toggle**: `hx-post`, swap a badge.
- **404/expired handling**: `HX-Redirect` response header for auth expiry mid-fragment.

### 10.3 Styling

Single `static/app.css`, hand-written, shadcn/ui look: neutral gray palette, CSS custom properties for radius/spacing/colors, components = cards, buttons, inputs, table, badge/tag (owned / missing / tradable / commander), dialog. Mobile-first, grid/flex — no framework. If it grows unwieldy → revisit Tailwind or templui (TD9).

Static files served via `go:embed` + `http.FileServer`, `Cache-Control: immutable` for fingerprinted assets in prod.

---

## 11. Observability

**MVP**: `slog` JSON to stdout (12-factor). Request-log middleware: method, path, status, duration, user id. Scryfall calls logged with cache-hit flag.

**Post-MVP**: OpenTelemetry SDK (`otlptracehttp`) → self-hosted or cloud SigNoz on the same VPS. Instrument `net/http` server, `database/sql` (otelsql), and the outbound Scryfall client. The tracing seam is added when needed — no premature instrumentation beyond the middleware pattern.

---

## 12. Deployment

- **Build**: `templ generate && sqlc generate && go build ./cmd/cockadeck` → one binary, assets + migrations embedded.
- **Run**: systemd unit on the VPS; `EnvironmentFile=/etc/cockadeck/env`; working dir holds `cockadeck.db`; restart `on-failure`.
- **TLS/proxy**: nginx on 443 → `localhost:8080`; certbot (`--nginx`) for the domain cert. HTTP→HTTPS redirect. *(TD7: Caddy would remove certbot entirely — single-file config, automatic TLS — but spec follows the stated certbot choice.)*
- **Backup**: `sqlite3 cockadeck.db ".backup ..."` cron, or litestream later. DB file is the whole state.
- **Deploys**: scp binary → `systemctl restart`. No CI initially; Makefile target `deploy` is enough at this scale.

---

## 13. Security Checklist

- Parameterized SQL only (sqlc enforces).
- Templ escapes output by default; never render raw Scryfall HTML.
- argon2id hashing; secrets via env; no secrets in logs.
- `Secure`/`HttpOnly`/`SameSite` cookies; Origin check on mutations.
- AuthZ by context-scoped queries (§6) — shared reads go through the `is_shared` predicate.
- Scryfall URLs/params URL-escaped; response size-bounded (`http.MaxBytesReader` on requests, sane client limits).
- Rate-limit middleware on `/login`/`/register` (simple token bucket per IP) to blunt brute force.

---

## 14. Testing Strategy

| Layer | Approach |
|---|---|
| sqlc/store | integration tests against a temp SQLite file (run goose migrations in test) — real DB, no mocks |
| services | fake `Querier` (sqlc emits the interface) for pure logic tests |
| handlers | `httptest` + real store; assert HTML fragments contain expected rows |
| scryfall | `httptest.Server` fixture responses; recorded JSON for a handful of cards |
| export | golden-file tests for `.cod` and CSV |

`Makefile`: `make test` = `templ generate && sqlc generate && go test ./...`.

---

## 15. Build Order (matches SPEC §2)

1. Skeleton: config, sqlite+goose, sqlc, auth (register/login/JWT middleware), layout.
2. Scryfall client + cache; set+number search; search page + fragments.
3. Decks CRUD, zones, add-from-search, deck view, filters.
4. `.cod` export. → **MVP deployable.**
5. Collections (entries, finishes, filters, set completion), CSV export.
6. Sharing toggle, shared-collection browser, tradable view.
7. Name search, other printings, stats, import, SigNoz — post-MVP.

---

## 16. Open Technical Decisions

| # | Decision | Options | Lean |
|---|---|---|---|
| TD1 | SQLite driver | `modernc.org/sqlite` (no cgo) vs `mattn/go-sqlite3` (faster, cgo) | modernc — simpler builds; swap is one line |
| TD2 | Auth tokens | JWT vs SQLite sessions | JWT per request; sessions arguably simpler — see §6 |
| TD3 | Migration tool | goose vs golang-migrate | goose — lighter, embeds well, sqlc-friendly |
| TD4 | User-facing identity | `display_name` required at register vs derive from email | required — needed for shared-collection browser |
| TD5 | PKs | INTEGER rowids vs UUIDs | INTEGER — SQLite-native; Scryfall IDs only for cards |
| TD6 | Password hashing | argon2id vs bcrypt | argon2id (OWASP); bcrypt acceptable fallback |
| TD7 | TLS | nginx+certbot vs Caddy | nginx+certbot per request |
| TD8 | Migration runner | goose at startup vs separate CLI step | at startup, embedded FS |
| TD9 | CSS | hand-written vs Tailwind/templui | hand-written now; revisit if it hurts |
| TD10 | HTMX version | vendored pinned `htmx.min.js` | pin latest 2.x, commit to `static/` |
