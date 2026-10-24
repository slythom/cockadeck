#!/bin/sh
# Cockadeck first-time VPS setup — run as root.
#
#   make deploy-bundle DEPLOY_HOST=root@vps.example.com
#   ssh root@vps 'DOMAIN=cockadeck.example.com sh /tmp/cockadeck-deploy/bootstrap.sh'
#
# Installs Caddy if missing, creates the service user, writes /etc/cockadeck/env
# (generating COCKADECK_JWT_SECRET on first run), installs the binary, unit,
# backup cron and Caddy site, then starts everything. Idempotent — safe to
# re-run; existing config is never overwritten.
set -eu

DOMAIN=${DOMAIN:?usage: DOMAIN=cockadeck.example.com sh bootstrap.sh}
BIN=${BIN:-/tmp/cockadeck}          # built binary shipped next to this script
HERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

[ -f "$BIN" ] || { echo "binary not found at $BIN — set BIN=/path/to/cockadeck"; exit 1; }

# --- system packages ---------------------------------------------------------

export DEBIAN_FRONTEND=noninteractive
command -v sqlite3 >/dev/null || apt-get install -y sqlite3

if ! command -v caddy >/dev/null; then
	apt-get install -y curl gpg debian-keyring debian-archive-keyring apt-transport-https
	curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
		| gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
	curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
		-o /etc/apt/sources.list.d/caddy-stable.list
	apt-get update
	apt-get install -y caddy
fi

# --- user + dirs -------------------------------------------------------------

id cockadeck >/dev/null 2>&1 || \
	useradd -r -d /var/lib/cockadeck -m -s /usr/sbin/nologin cockadeck
install -d -o cockadeck -g cockadeck /var/lib/cockadeck /var/backups/cockadeck
install -d -m700 -o cockadeck -g cockadeck /etc/cockadeck

# --- env (first run only) ------------------------------------------------------

if [ ! -f /etc/cockadeck/env ]; then
	install -m600 -o cockadeck -g cockadeck "$HERE/env.example" /etc/cockadeck/env
	secret=$(head -c32 /dev/urandom | od -An -tx1 | tr -d ' \n')
	sed -i "s/^COCKADECK_JWT_SECRET=\$/COCKADECK_JWT_SECRET=$secret/" /etc/cockadeck/env
	echo "generated COCKADECK_JWT_SECRET in /etc/cockadeck/env"
fi

# --- binary, unit, backup cron -------------------------------------------------

install -m755 "$BIN" /usr/local/bin/cockadeck
install -m644 "$HERE/cockadeck.service" /etc/systemd/system/cockadeck.service
install -m755 "$HERE/cockadeck-backup.sh" /usr/local/bin/cockadeck-backup
printf '0 4 * * * root /usr/local/bin/cockadeck-backup\n' > /etc/cron.d/cockadeck-backup

# --- caddy site ----------------------------------------------------------------

if ! grep -q "^$DOMAIN " /etc/caddy/Caddyfile 2>/dev/null; then
	printf '\n%s {\n\treverse_proxy 127.0.0.1:8080\n\tencode zstd gzip\n}\n' \
		"$DOMAIN" >> /etc/caddy/Caddyfile
fi

systemctl daemon-reload
systemctl enable --now cockadeck
systemctl reload caddy

echo
echo "Done. https://$DOMAIN"
echo "  app:    systemctl status cockadeck · journalctl -u cockadeck -f"
echo "  backup: /var/backups/cockadeck (daily 04:00)"
