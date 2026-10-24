# cockadeck

MTG deck builder and collection tracker. Server-rendered Go + HTMX + SQLite,
card data via the Scryfall API. See `SPEC.md` (features), `TECHNICAL_SPEC.md`
(architecture) and `PLAN.md` (roadmap).

## Development

Requires Go 1.26+ (templ and sqlc run via `go tool`).

```sh
make generate   # templ + sqlc codegen
make dev        # run on :8080 with a dev JWT secret
make test       # codegen + go test ./...
make build      # ./cockadeck binary
```

## Deployment

Single static binary behind a Caddy reverse proxy (automatic TLS). Files live
in `deploy/`:

- `deploy/bootstrap.sh` — first-time server setup (user, env, unit, cron, Caddy)
- `deploy/Caddyfile` — site block; bootstrap appends it for your domain
- `deploy/cockadeck.service` — systemd unit
- `deploy/env.example` — `COCKADECK_*` vars, becomes `/etc/cockadeck/env`
- `deploy/cockadeck-backup.sh` — daily `sqlite3 .backup` snapshot (cron)

First-time setup — point the domain's DNS at the VPS first, then:

```sh
make deploy-bundle DEPLOY_HOST=root@vps.example.com
ssh root@vps.example.com 'DOMAIN=cockadeck.example.com sh /tmp/cockadeck-deploy/bootstrap.sh'
```

Bootstrap is idempotent: it installs Caddy if absent, creates the `cockadeck`
user, writes `/etc/cockadeck/env` with a generated `COCKADECK_JWT_SECRET`,
installs the unit + backup cron, registers the Caddy site, and starts the
service. Caddy issues the TLS cert on first request.

Ship a new release:

```sh
make deploy DEPLOY_HOST=user@vps.example.com
```

