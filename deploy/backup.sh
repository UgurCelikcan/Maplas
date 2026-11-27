#!/usr/bin/env bash
# Maplas - veritabanı ve yüklenen resimlerin yedeği.
# Yedekler /var/backups/maplas içine yazılır, 14 günden eski olanlar silinir.
#
# Elle çalıştırmak için:  /opt/maplas/deploy/backup.sh
# Geri yüklemek için:
#   gunzip -c /var/backups/maplas/db-TARIH.sql.gz | docker compose exec -T db psql -U maplas_user maplas_db

set -euo pipefail

APP_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/maplas}"
KEEP_DAYS="${KEEP_DAYS:-14}"
STAMP="$(date +%Y-%m-%d_%H%M)"

cd "$APP_DIR"
mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"

env_value() { grep -E "^$1=" .env | head -n1 | cut -d= -f2- | sed 's/[[:space:]]*#.*$//' | tr -d '"'; }
DB_USER="$(env_value DB_USER)"
DB_NAME="$(env_value DB_NAME)"

echo "[$(date)] Yedek başlıyor"

docker compose exec -T db pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_DIR/db-$STAMP.sql.gz"

TMP="$(mktemp -d)"
docker compose cp backend:/app/uploads "$TMP/uploads"
tar -czf "$BACKUP_DIR/uploads-$STAMP.tar.gz" -C "$TMP" uploads
rm -rf "$TMP"

find "$BACKUP_DIR" -type f -mtime +"$KEEP_DAYS" -delete

echo "[$(date)] Yedek tamam: $BACKUP_DIR"
