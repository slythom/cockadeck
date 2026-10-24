#!/bin/sh
# Cockadeck DB backup — install to /usr/local/bin/cockadeck-backup, then cron:
#   0 4 * * * /usr/local/bin/cockadeck-backup
# The .backup command is snapshot-safe against the live WAL database.
# Keeps the last 14 daily snapshots.
set -eu

DB=/var/lib/cockadeck/cockadeck.db
DEST=/var/backups/cockadeck

mkdir -p "$DEST"
sqlite3 "$DB" ".backup '$DEST/cockadeck-$(date +%F).db'"

ls -1t "$DEST"/cockadeck-*.db 2>/dev/null | tail -n +15 | xargs -r rm --
