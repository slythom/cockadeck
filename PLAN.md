# Cockadeck — Implementation Plan

| | |
|---|---|
| Status | Working plan, derived from `SPEC.md` + `TECHNICAL_SPEC.md` vs. current code |
| Date | 2026-10-24 |

---

## 1. Current State

The codebase already covers TECHNICAL_SPEC §15 build steps 1–4 — **the MVP path is functionally complete**: auth → set+number search → add to deck → deck view → `.cod` export.

| Area | Status | Evidence |
|---|---|---|
| Auth (FR-AUTH-01–05) | Done | `internal/auth/` (argon2id, JWT, middleware), `auth_handlers.go`, IP rate-limit on auth routes, CSRF origin check, tests |
| Config / DB / migrations / sqlc | Done | `internal/config`, `internal/store/db.go` (WAL, pragmas, embedded goose), `db/migrations/00001_init.sql` — full schema including `collections`/`collection_cards` |
| Scryfall client + cache (FR-SEARCH-01–04) | Done | `internal/scryfall/scryfall.go` (75 ms limiter, error mapping), cache-fill in `catalog.CardBySetNumber` |
| Search page + fragments | Done (set+num only) | `handlers_search.go`, `views/search.templ` |
| Decks (FR-DECK-01–06) | Done | `handlers_decks.go`, `catalog.go` deck methods, zone-merge transaction, filters (set/rarity/cmc), `views/deck.templ` |
| `.cod` export (FR-EXP-01/02/04) | Done | `internal/export/export.go` + golden test, `GET /decks/{id}/export.cod` |
| Middleware chain | Done | recover → requestLog → secureHeaders → CSRF → RequireAuth |
| Collection (FR-COLL-01–05) | **Not started** | schema exists; no `db/queries/collections.sql`, no service methods, handlers, or views |
| CSV export (FR-EXP-03) | **Not started** | — |
| Sharing + tradable view (FR-CMP-01–05) | **Not started** | — |
| Name search (FR-SEARCH-06) | Done | `SearchByName` in scryfall client + `CardsByName` cache-fill; name field is optional alongside set+num, live results on keyup |
| Other printings (FR-SEARCH-07, FF) | **Not started** | — |
| In-deck marker in results (FR-DECK-07, FF) | **Not started** | — |
| Stats / import / OTel (Post-MVP) | Not started | deferred by design |
| Deployment artifacts | Done (pending a host) | `deploy/` — Caddyfile, systemd unit, env example, backup script; `make deploy`; README rewritten |

Note: `go` is not installed on this machine, so `make test` could not be verified — the last commits were presumably tested elsewhere.

---

## 2. Proposed Phases

### Phase A — MVP gaps & fast-follows (small, high value)

1. ~~**Name search (FR-SEARCH-06)**~~ — **done 2026-10-24**: `GET /cards/search` client, list fragment, keyup-triggered live results, set code usable as an optional narrowing constraint. Pulled into MVP by decision below.
2. **In-deck marker (FR-DECK-07)** — join search results against the user's `deck_cards`; show a badge on printings already in a deck.
3. **Other printings (FR-SEARCH-07)** — `unique:prints` query from a card result; list fragment linking back into set+num lookups (each printing lands in cache for free).
4. **Commander deck-size soft indicator** (§7.4) — show "97/100" next to the commander-deck card total. Cheap.
5. **Small fixes** — tighten `.gitignore` (`cockadeck` → `/cockadeck`; the bare pattern also matches the `cmd/cockadeck/` directory), flesh out README (run/dev instructions).

### Phase B — Increment 2: Collection (FR-COLL, FR-EXP-03)

Build order inside the phase:

1. **Queries** — new `db/queries/collections.sql`: `GetOrCreateDefaultCollection` (upsert on `user_id`), `UpsertCollectionCard` (adds quantity on conflict, same pattern as `UpsertDeckCard`), `SetCollectionCardQuantity`, `DeleteCollectionCard`, `ListCollectionCards` with `sqlc.narg` filters (set / rarity / cmc / finish), `ListDistinctSetCodes` for the filter dropdown, `SetCollectionShared`.
   - Schema note: `collections` has no `UNIQUE(user_id)` — either add one in a new migration (recommended; makes get-or-create race-safe) or keep the read-then-insert pattern.
2. **Service** (`catalog.go`) — `CollectionCards(ctx, userID, filters)`, `AddCollectionCard`, `UpdateCollectionCard`, `RemoveCollectionCard`, `ToggleSharing`. Mirror the deck-card methods; finish becomes part of the row key like zone.
3. **Search-result "add" UI** — extend the add popover with a collection target (qty + finish select), per SPEC §7.5/D2: one result, both deck and collection targets. Decide UI shape — see §3 Q2.
4. **Collection page** — `GET /collection`, own-collection view reusing the deck-row markup pattern: image, name, set, #, rarity, mana cost, finish, qty +/−, remove. Filters as query params + HTMX fragment refresh (`hx-push-url`), same as the deck page. Add "Collection" to the nav.
5. **Set completion (FR-COLL-05)** — when `?set=<code>` is present:
   - Scryfall client gains `SetCards(ctx, code)` — paginated `GET /cards/search?q=set:{code} unique:prints&order=collector_number`, following `has_more`/`next_page`, upserting each card into `cards`.
   - View then `LEFT JOIN`s the full set against `collection_cards` → **owned** / **missing** badge per printing. Filter by owned/missing too (FR-COLL-04).
   - Cache note: re-crawling a set on every view is wasteful — check a "set already cached" heuristic (e.g. count of cached cards for the set > 0 + `fetched_at` freshness, or a `sets` table with `synced_at`). Decide TTL — see §3 Q3.
6. **CSV export (FR-EXP-03)** — `GET /collection/export.csv`, `encoding/csv`, columns `name,set_code,set_name,collector_number,rarity,finish,quantity`, filename `collection-YYYY-MM-DD.csv`. Golden-file test next to `deck.cod`.

### Phase C — Increment 2+: Sharing & tradable view (FR-CMP)

1. **Sharing toggle (FR-CMP-01)** — `POST /collection/sharing` flipping `collections.is_shared`; badge swap on the collection page.
2. **Shared-collection browser (FR-CMP-02)** — `GET /collections` listing users (display_name) whose collection is shared. Nav entry.
3. **Read-only shared view (FR-CMP-03/05)** — `GET /collections/{userID}` rendering the owner's rows through the spec's `is_shared`-in-SQL query (TECHNICAL_SPEC §5.4) with the `tradable` flag computed by the name-level `NOT EXISTS` subquery. Same filters + a tradable-status filter (FR-CMP-04). No mutation controls rendered; handler has no write endpoints.
4. **Privacy re-check (NFR-04)** — tests: private collection returns 404/zero rows even when hitting the URL directly; cross-user writes impossible.

### Phase D — Deployment (needed for "hosted on the web") — **artifacts done 2026-10-24**

1. ~~`deploy/` — systemd unit, Caddyfile, env example, backup script~~ — written.
2. ~~Backup: `deploy/cockadeck-backup.sh` (cron `sqlite3 .backup`, 14-day retention)~~ — written.
3. ~~`make deploy DEPLOY_HOST=...` (cross-compile → scp → restart)~~ — written.
4. ~~README: dev setup + first-time server setup~~ — written.
5. ~~App hardening~~ — graceful shutdown on SIGTERM (`main.go`), `Cache-Control` on `/static/`, favicon. Done.
6. ~~First-time setup automation~~ — `deploy/bootstrap.sh` (idempotent: installs Caddy, user, env+generated secret, unit, backup cron, site; starts everything). Done.
7. **Remaining:** pick the VPS + domain, then `make deploy-bundle` + `bootstrap.sh` — the only step left needs the actual host (see §3 Q7).

### Phase E — Post-MVP (parked, order TBD)

- Deck statistics (FR-STAT-01–03): mana curve chart (inline SVG bars are enough — no chart lib), avg MV, type distribution, land guidance.
- Import (FR-IMP-01–02): plaintext list first; `.cod`/CSV later.
- OpenTelemetry → SigNoz.
- Possible extras: `etched` finish, card-price display (data already in `raw` JSON), cache TTL/refresh.

---

## 3. Open Questions & Decisions

**Q1. Pull name search into the current version?** — **Decided 2026-10-24: yes, done.**
Name search ships in the current version alongside set+number: a name alone searches Scryfall, set code optionally narrows it, set+num alone still does the exact lookup.

**Q2. Add-to-collection UX shape.**
SPEC §7.5 wants deck + collection targets on the same result. Options: (a) extend the current `<details>` popover with a second section (qty + finish + "Add to collection"), (b) a single unified form with two checkboxes/targets. (a) is simpler and matches what exists; (b) is one click fewer but more code. Lean: (a).

**Q3. Set-completion caching policy.** — **Decided 2026-10-24: crawl-once, trust the cache.**
First set view crawls Scryfall (paginated), fills `cards`, and subsequent views serve purely from cache. Printings are immutable so no TTL; revisit only if staleness ever shows up.

**Q4. TD7 — nginx+certbot vs Caddy.** — **Decided 2026-10-24: Caddy.**
Automatic ACME cert + renewal, ~5-line config, one fewer daemon. Artifacts written: `deploy/Caddyfile`, `deploy/cockadeck.service`, `deploy/env.example`, `deploy/cockadeck-backup.sh`, `make deploy`.

**Q5. TD2 — JWT vs sessions.** — **Decided 2026-10-24: keep JWT.**
No changes required for deployment; stateless tokens stay.

**Q6. Multiple collections.** — **Decided 2026-10-24: single "Default" collection in the UI.**
Schema still supports N per user; no collection picker anywhere.

**Q7. Deployment target details.** — *Still open for Phase D*: VPS provider/OS, domain name, TLS termination point.

---

## 4. Suggested immediate next step

Phase B items 1–2 (collection queries + service), then the collection page — or Phase A remainder (in-deck marker, other printings) if you want to finish polishing search first.
