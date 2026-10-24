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

- `deploy/Caddyfile` — site block; set your domain, Caddy handles the cert
- `deploy/cockadeck.service` — systemd unit (`/etc/systemd/system/`)
- `deploy/env.example` — `COCKADECK_*` vars, goes to `/etc/cockadeck/env`
- `deploy/cockadeck-backup.sh` — daily `sqlite3 .backup` snapshot (cron)

First-time server setup:

```sh
useradd -r -d /var/lib/cockadeck -m cockadeck
install -m600 -o cockadeck deploy/env.example /etc/cockadeck/env   # edit JWT_SECRET
install -m755 deploy/cockadeck-backup.sh /usr/local/bin/cockadeck-backup
# append deploy/Caddyfile to /etc/caddy/Caddyfile with your domain
systemctl enable --now cockadeck && systemctl reload caddy
```

Ship a new release:

```sh
make deploy DEPLOY_HOST=user@vps.example.com
```
