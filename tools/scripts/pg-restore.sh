#!/usr/bin/env bash
# pg-restore.sh — Download a PostgreSQL backup from S3/MinIO and restore it.
#
# Usage:
#   ./tools/scripts/pg-restore.sh <dump-filename>
#
#   <dump-filename>  The .sql.gz filename inside BACKUP_PREFIX (e.g. pulzifi_20240115T030000Z.sql.gz).
#                    Run 'aws s3 ls s3://<bucket>/<prefix>/' to list available dumps.
#
# Required env vars (same names as shared/config/config.go):
#   DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD
#   MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY
#
# Optional env vars:
#   BACKUP_BUCKET    S3/MinIO bucket for backups (default: pulzifi-backups)
#   BACKUP_PREFIX    Key prefix inside bucket   (default: postgres/daily)
#   PSQL_EXTRA_OPTS  Extra flags forwarded to psql (default: empty)
#
# !!WARNING!! This DROPS and RECREATES the target database before restoring.
# A confirmation prompt is shown unless FORCE_RESTORE=1 is set.
#
# Dependencies: psql, gunzip (gzip), aws CLI
#
# Exit codes:
#   0  success
#   1  configuration / argument error
#   2  download failed
#   3  restore failed
#   4  user aborted

set -euo pipefail

# ── helpers ───────────────────────────────────────────────────────────────────

LOG_PREFIX="[pg-restore]"

log()  { echo "${LOG_PREFIX} $(date -u '+%Y-%m-%dT%H:%M:%SZ') INFO  $*"; }
warn() { echo "${LOG_PREFIX} $(date -u '+%Y-%m-%dT%H:%M:%SZ') WARN  $*" >&2; }
die()  { echo "${LOG_PREFIX} $(date -u '+%Y-%m-%dT%H:%M:%SZ') ERROR $*" >&2; exit "${1}"; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die 1 "Required command '$1' not found. Install it before running this script."
}

require_var() {
  local val
  val="${!1:-}"
  [ -n "$val" ] || die 1 "Required environment variable '$1' is not set."
}

# ── argument check ─────────────────────────────────────────────────────────────

if [ "${#}" -lt 1 ] || [ -z "${1:-}" ]; then
  echo "Usage: $0 <dump-filename>"
  echo ""
  echo "  <dump-filename>  e.g. pulzifi_20240115T030000Z.sql.gz"
  echo ""
  echo "List available dumps:"
  echo "  aws s3 ls s3://\${BACKUP_BUCKET:-pulzifi-backups}/\${BACKUP_PREFIX:-postgres/daily}/"
  exit 1
fi

DUMP_FILENAME="$1"

# ── dependency checks ─────────────────────────────────────────────────────────

require_cmd psql
require_cmd gzip
require_cmd aws

# ── configuration ─────────────────────────────────────────────────────────────

require_var DB_HOST
require_var DB_PORT
require_var DB_NAME
require_var DB_USER
require_var DB_PASSWORD
require_var MINIO_ENDPOINT
require_var MINIO_ACCESS_KEY
require_var MINIO_SECRET_KEY

BACKUP_BUCKET="${BACKUP_BUCKET:-pulzifi-backups}"
BACKUP_PREFIX="${BACKUP_PREFIX:-postgres/daily}"
# Default ON_ERROR_STOP=1 so that any SQL error mid-restore causes psql to exit
# non-zero immediately rather than silently continuing on a partially-restored DB.
PSQL_EXTRA_OPTS="${PSQL_EXTRA_OPTS:--v ON_ERROR_STOP=1}"
FORCE_RESTORE="${FORCE_RESTORE:-0}"

# Derive MinIO endpoint URL.
MINIO_URL="${MINIO_ENDPOINT%/}"
case "$MINIO_URL" in
  http://*|https://*) ;;
  *) MINIO_URL="http://${MINIO_URL}" ;;
esac

export AWS_ACCESS_KEY_ID="$MINIO_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="$MINIO_SECRET_KEY"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"
export AWS_ENDPOINT_URL="$MINIO_URL"
export AWS_PAGER=""

DUMP_KEY="${BACKUP_PREFIX}/${DUMP_FILENAME}"

# ── confirmation guard ────────────────────────────────────────────────────────

echo ""
echo "  ┌─────────────────────────────────────────────────────────┐"
echo "  │              !! DESTRUCTIVE OPERATION !!                │"
echo "  │                                                         │"
echo "  │  This will DROP and RECREATE the database:              │"
echo "  │    Database : ${DB_NAME}"
echo "  │    Host     : ${DB_HOST}:${DB_PORT}"
echo "  │    Dump     : ${DUMP_FILENAME}"
echo "  │                                                         │"
echo "  │  ALL CURRENT DATA IN THIS DATABASE WILL BE LOST.       │"
echo "  └─────────────────────────────────────────────────────────┘"
echo ""

if [ "${FORCE_RESTORE}" != "1" ]; then
  read -r -p "Type the database name to confirm restore [${DB_NAME}]: " CONFIRM
  if [ "$CONFIRM" != "$DB_NAME" ]; then
    die 4 "Confirmation did not match '${DB_NAME}'. Aborting."
  fi
else
  warn "FORCE_RESTORE=1 is set — skipping confirmation prompt."
fi

# ── download ──────────────────────────────────────────────────────────────────

DUMP_TMPFILE="$(mktemp "/tmp/pg-restore-XXXXXX.sql.gz")"
trap 'rm -f "$DUMP_TMPFILE"' EXIT

log "Downloading s3://${BACKUP_BUCKET}/${DUMP_KEY} ..."

if ! aws s3 cp "s3://${BACKUP_BUCKET}/${DUMP_KEY}" "$DUMP_TMPFILE"; then
  die 2 "Download failed. Check MINIO_ENDPOINT, credentials, bucket '${BACKUP_BUCKET}', and key '${DUMP_KEY}'."
fi

DUMP_SIZE="$(du -h "$DUMP_TMPFILE" | cut -f1)"
log "Download complete — compressed size: ${DUMP_SIZE}"

# ── restore ───────────────────────────────────────────────────────────────────

export PGPASSWORD="$DB_PASSWORD"

# Connect to the default 'postgres' maintenance database to drop/recreate the target.
PSQL_ADMIN="psql \
  --host=$DB_HOST \
  --port=$DB_PORT \
  --username=$DB_USER \
  --dbname=postgres \
  --no-password"

# ── dump integrity check ──────────────────────────────────────────────────────
# Validate the gzip file BEFORE dropping the database. A corrupt or truncated
# archive must abort here so we never reach the point-of-no-return DROP step.

log "Verifying dump integrity (gzip -t)..."
if ! gzip -t "$DUMP_TMPFILE"; then
  die 3 "Dump integrity check failed — '${DUMP_FILENAME}' is corrupt or truncated. Aborting (database NOT dropped)."
fi
log "Dump integrity OK."

log "Terminating active connections to '${DB_NAME}'..."
$PSQL_ADMIN -c "
  SELECT pg_terminate_backend(pid)
  FROM pg_stat_activity
  WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();
" > /dev/null 2>&1 || true

log "Dropping database '${DB_NAME}'..."
$PSQL_ADMIN -c "DROP DATABASE IF EXISTS \"${DB_NAME}\";" \
  || die 3 "Could not drop database '${DB_NAME}'. Check superuser permissions."

log "Recreating database '${DB_NAME}'..."
$PSQL_ADMIN -c "CREATE DATABASE \"${DB_NAME}\";" \
  || die 3 "Could not recreate database '${DB_NAME}'."

log "Restoring dump into '${DB_NAME}'..."

# shellcheck disable=SC2086
if ! gzip -dc "$DUMP_TMPFILE" | psql \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="$DB_NAME" \
    --no-password \
    $PSQL_EXTRA_OPTS \
    > /dev/null; then
  unset PGPASSWORD
  die 3 "psql restore failed. Check the dump file integrity and DB credentials."
fi

unset PGPASSWORD

log "Restore complete. Database '${DB_NAME}' is now live from dump '${DUMP_FILENAME}'."
log "Run 'make migrate' if schema migrations need to be applied on top of the restored data."

exit 0
