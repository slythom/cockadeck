# Cockadeck — Feature Specification

| | |
|---|---|
| Status | Draft v1 |
| Date | 2026-09-16 |
| Audience | Maintainer(s) of Cockadeck |
| Project type | Learning / personal project, shared with friends |

---

## 1. Purpose

Cockadeck is a web application for **Magic: The Gathering** players to:

1. Look up cards via the Scryfall API.
2. Build **decks** (primary use case) and track a **card collection** (secondary).
3. Export decks to the Cockatrice `.cod` format and collections to CSV.
4. (Post-MVP) Visualize deck statistics: mana curve, type distribution, land guidance.

The app is authentication-gated, hosted on the web, responsive, and designed for a fast, low-latency interface (no full page reloads for common interactions).

---

## 2. Scope & Priorities

| Priority | Meaning |
|---|---|
| **MVP** | First usable release. |
| **FF** | Fast follow — included in MVP if cheap, otherwise immediately after. |
| **Post-MVP** | Deliberately deferred; schema/design must not block it. |
| **Out** | Explicitly out of scope. |

### Release ordering (deck-first)

1. **MVP**: Auth → card search → add to deck → deck view → `.cod` export.
2. **Increment 2**: Collection (add, view, filter) → CSV export → collection comparison (tradable view).
3. **Post-MVP**: Name search, other printings, deck statistics, import.

---

## 3. Definitions

| Term | Meaning |
|---|---|
| **Card / printing** | A specific physical printing identified by Scryfall ID — i.e. (set code, collector number). "Lightning Bolt, M11 #146". |
| **Card name** | The printing-agnostic identity of a card ("Lightning Bolt"). Used for owned/missing matching. |
| **Finish** | `nonfoil` or `foil`. A collection entry is per (printing, finish). |
| **Collection** | A user's inventory of owned printings, with finish and quantity. |
| **Deck** | A named list of cards organized in zones, owned by a user. |
| **Zone** | `main` (default), `sideboard`, or `commander`. |
| **Format** | Deck metadata: `40-card` or `commander`. Informational only — no validation in MVP. |
| **Owned / missing** | A collection/set-completion concept, per **printing**: when viewing the contents of a set, a card is "owned" if the user has it in their collection, otherwise "missing" (e.g. #10 and #147 missing from OTJ). Not shown in decks. |
| **Tradable** | A comparison tag, per **card name**: when viewing another user's shared collection, a card is tradable if the viewer owns zero copies of that card name (any printing, any finish). Deliberately different granularity from owned/missing. |
| **Shared collection** | A collection whose owner has opted in to making it visible to other users. Default: private. |

---

## 4. Actors

- **Visitor** — unauthenticated; can only reach the registration/login pages.
- **User** — authenticated; manages their own decks and collections.
- No admin role, no sharing between users in this version.

---

## 5. Functional Requirements

Requirements use IDs (`FR-<area>-<n>`) for traceability.

### 5.1 Authentication (MVP)

| ID | Requirement | Priority |
|---|---|---|
| FR-AUTH-01 | A visitor can self-register with email + password. | MVP |
| FR-AUTH-02 | A user can log in and log out. Sessions persist across browser restarts. | MVP |
| FR-AUTH-03 | Every application page except login/register requires authentication. | MVP |
| FR-AUTH-04 | Passwords are stored hashed (never plaintext, never logged). | MVP |
| FR-AUTH-05 | Registration validates email format and a minimum password strength; duplicate emails are rejected with a clear message. | MVP |
| — | Two-factor authentication | Out |
| — | Password reset / email verification | Out (see §7) |

### 5.2 Card Search (MVP)

| ID | Requirement | Priority |
|---|---|---|
| FR-SEARCH-01 | A user can search for a card by **set code + collector number**. The lookup resolves against the Scryfall API. | MVP |
| FR-SEARCH-02 | Results display at minimum: card image, name, set name, collector number, rarity. | MVP |
| FR-SEARCH-03 | Non-ASCII / special collector numbers (e.g. `361★`, `5a`) are supported. | MVP |
| FR-SEARCH-04 | Clear, distinct messages for: card not found, invalid input, Scryfall unavailable. | MVP |
| FR-SEARCH-05 | Each result offers "add" actions targeting a deck and/or a collection (see 5.3, 5.4). | MVP |
| FR-SEARCH-06 | A user can search by card name (Scryfall `/cards/search`), returning a list of matching cards. | FF |
| FR-SEARCH-07 | From a result, a user can see other printings of the same card. | FF |

### 5.3 Decks (MVP)

| ID | Requirement | Priority |
|---|---|---|
| FR-DECK-01 | A user can create, rename, edit description, and delete decks. | MVP |
| FR-DECK-02 | A deck has: name (required), description (optional), format (`40-card` or `commander`). | MVP |
| FR-DECK-03 | From a search result, a user can add a card to a chosen deck, choosing quantity and zone (`main` by default, `sideboard`, `commander` for commander-format decks). | MVP |
| FR-DECK-04 | Within a deck, a user can change a card's quantity, move it between zones, and remove it. | MVP |
| FR-DECK-05 | Deck view lists cards with: image, name, set name, collector number, rarity, mana cost. Cards in the `commander` zone are visually marked (label/tag). | MVP |
| FR-DECK-06 | Deck view can be filtered by: set (name or code), rarity, mana cost. Filters can be combined. | MVP |
| FR-DECK-07 | Cards already in the user's decks are visually distinguishable in search results. | FF |
| — | Format legality validation (singleton, deck size, banned lists) | Out (see §7) |

### 5.4 Collection (Increment 2)

| ID | Requirement | Priority |
|---|---|---|
| FR-COLL-01 | From a search result, a user can add a card to their collection with quantity and finish (`nonfoil` default, `foil`). | MVP (Incr. 2) |
| FR-COLL-02 | A user can increment/decrement quantity and remove collection entries. Entries are per (printing, finish). | MVP (Incr. 2) |
| FR-COLL-03 | Collection view lists entries with: image, name, set name, collector number, rarity, mana cost, finish, quantity. | MVP (Incr. 2) |
| FR-COLL-04 | Collection view can be filtered by: set (name or code), rarity, mana cost, finish, owned/missing. Filters can be combined. | MVP (Incr. 2) |
| FR-COLL-05 | **Set completion**: when filtered to a set, the collection view displays *all* cards of that set (fetched from Scryfall), each marked **owned** or **missing** per printing. | MVP (Incr. 2) |

### 5.5 Collection Comparison ("Tradable" view) (Incr. 2+)

Depends on 5.4. Lets a user find trade candidates in other users' collections.

| ID | Requirement | Priority |
|---|---|---|
| FR-CMP-01 | A user can toggle their collection's visibility to other users (shared / private). Default: private. | Incr. 2+ |
| FR-CMP-02 | A user can browse a list of users who share their collection and open one. Private collections are not listed or reachable. | Incr. 2+ |
| FR-CMP-03 | Viewing another user's shared collection shows the same fields and filters as one's own collection view, plus a **Tradable** tag (or equivalent visual overlay) on every card the viewer owns zero copies of by card name (any printing, any finish). | Incr. 2+ |
| FR-CMP-04 | The comparison view can be filtered by tradable status. | Incr. 2+ |
| FR-CMP-05 | The comparison view is read-only: the viewer cannot modify the other user's collection. | Incr. 2+ |
| — | Reverse comparison ("what I own that they miss"), trade offers, messaging between users | Out |

### 5.6 Export (MVP)

| ID | Requirement | Priority |
|---|---|---|
| FR-EXP-01 | A user can export a deck as a `.cod` file (Cockatrice XML): zones `main` and `side`, each `<card>` carrying `name` and `number` (quantity). Export is identical regardless of deck format — no commander-specific `.cod` handling; in-app `commander`-zone cards map to `main`. | MVP |
| FR-EXP-02 | The `.cod` file includes the deck name (and description as comments). | MVP |
| FR-EXP-03 | A user can export their collection as CSV — one row per (printing, finish) with at least: name, set code, set name, collector number, rarity, finish, quantity. | MVP (Incr. 2) |
| FR-EXP-04 | Exported files download with sensible filenames (e.g. `deckname.cod`, `collection-2026-09-16.csv`). | MVP |

### 5.7 Deck Statistics (Post-MVP)

| ID | Requirement | Priority |
|---|---|---|
| FR-STAT-01 | Mana curve chart of a deck's main zone. | Post-MVP |
| FR-STAT-02 | Average mana value, land/spell breakdown, card-type distribution. | Post-MVP |
| FR-STAT-03 | Recommended land count by color (heuristic TBD). | Post-MVP |

### 5.8 Import (Post-MVP)

| ID | Requirement | Priority |
|---|---|---|
| FR-IMP-01 | Import a plaintext decklist (`4 Lightning Bolt` lines) into a deck. | Post-MVP |
| FR-IMP-02 | Import `.cod` / CSV files. | Post-MVP |

---

## 6. Non-Functional Requirements

| ID | Requirement |
|---|---|
| NFR-01 | **Reactivity**: common interactions (search, add, filter, quantity change) update the UI without full page reloads; the interface must feel low-latency on a normal connection. |
| NFR-02 | **Responsive**: usable on desktop and mobile widths; primary flows (search → add → view → export) work on a phone. |
| NFR-03 | **UI style**: component look-and-feel inspired by shadcn/ui (grayscale palette, consistent radius/spacing, cards, dialogs, tables, form controls). |
| NFR-04 | **Privacy**: users can only *modify* their own data. A user's collection is *visible* to other users only if they explicitly enabled sharing (FR-CMP-01); private collections and non-collection data are never exposed. |
| NFR-05 | **Scryfall etiquette**: card data fetched live is cached locally; requests respect Scryfall's rate-limit guidance (~10 req/s, 50–100 ms spacing). Card images are hotlinked from Scryfall's CDN. |
| NFR-06 | **Hosting**: deployable as a single self-contained web application on a small VPS/PaaS; no external service dependencies beyond Scryfall. |

---

## 7. Open Questions & Assumptions

Assumptions recorded during design — flag if any are wrong:

1. **Password reset**: out of scope; for a friends-scale deployment, manual reset by the operator is acceptable.
2. **Collections count**: data model supports multiple named collections per user; MVP UI only needs to expose one default collection.
3. **Finish values**: `nonfoil`/`foil` only. `etched` exists on Scryfall and can be added later without schema pain.
4. **Deck format validation**: `40-card`/`commander` is stored metadata only; no deck-size or singleton enforcement in MVP. A soft "your commander deck has 97 cards" indicator could be a cheap FF if wanted.
5. **Adding to both deck and collection**: the add UI on a result allows deck + collection targets simultaneously (per Q2), not as two separate flows.

---

## 8. Decision Log

| # | Decision | Source |
|---|---|---|
| D1 | Cards are printing-specific (Scryfall ID); collection entries also carry finish. | Q1 |
| D2 | Decks and collections are independent lists; a result can be added to both. | Q2 |
| D3 | Quantities everywhere; deck zones = main / sideboard / commander. | Q3, Q10 |
| D4 | Set+collector-number is the core lookup; name search fast-follow. | Q4 |
| D5 | Open self-registration, email+password, no 2FA. | Q5 |
| D6 | "Reactive" = fast/low-latency UI, not a specific tech stack. | Q6 |
| D7 | Deck formats: `40-card` and `commander` only. | Q10 |
| D8 | MVP is deck-first: search → deck → `.cod`; collection + CSV next; stats last. | Q14 |
| D9 | Import deferred post-MVP but schema must not block it. | Q11 |
| D10 | Live Scryfall API calls with local caching. | Q9 |
| D11 | Owned/missing is a collection set-completion concept (per printing); decks do not show it. | §7.1 |
| D12 | No commander-specific `.cod` handling; in-app `commander` zone exports to `main`. Commander is shown as a UI label/tag only. | §7.2 |
| D13 | Collection visibility is opt-in per user (default private); only shared collections are browsable. | §5.5 |
| D14 | Tradable is name-level (viewer owns zero copies of the card name); comparison is one-way (viewer → other's collection). | §5.5 |
