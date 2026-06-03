#!/usr/bin/env bash
# pg-backup.sh — Daily PostgreSQL backup: pg_dump → gzip → upload to S3/MinIO.
#
# Usage:
#   ./tools/scripts/pg-backup.sh
#
# Required env vars (same names as shared/config/config.go):
#   DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD
#   MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY
#
# Optional env vars:
#   BACKUP_BUCKET          S3/MinIO bucket for backups (default: pulzifi-backups)
#   BACKUP_PREFIX          Key prefix inside bucket   (default: postgres/daily)
#   BACKUP_RETENTION_COUNT  How many most-recent dumps to keep (default: 30)
#   PG_DUMP_EXTRA_OPTS     Extra flags forwarded to pg_dump (default: empty)
#
# Dependencies: pg_dump, gzip, aws CLI (configured to point at MinIO via AWS_ENDPOINT_URL)
#
# Exit codes:
#   0  success
#   1  configuration error (missing required var or tool)
#   2  pg_dump failed
#   3  upload failed
#   4  retention pruning failed (non-fatal — backup itself succeeded)

set -euo pipefail

# ── helpers ──────────────────────────────────────────────────────────────────

LOG_PREFIX="[pg-backup]"

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

# ── dependency checks ─────────────────────────────────────────────────────────

require_cmd pg_dump
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
BACKUP_RETENTION_COUNT="${BACKUP_RETENTION_COUNT:-30}"  # keep N most recent dumps
PG_DUMP_EXTRA_OPTS="${PG_DUMP_EXTRA_OPTS:-}"

# Derive MinIO endpoint URL — strip trailing slash.
MINIO_URL="${MINIO_ENDPOINT%/}"
# Ensure scheme is present.
case "$MINIO_URL" in
  http://*|https://*) ;;
  *) MINIO_URL="http://${MINIO_URL}" ;;
esac

# Export AWS-compatible credentials for the aws CLI.
export AWS_ACCESS_KEY_ID="$MINIO_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="$MINIO_SECRET_KEY"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"
export AWS_ENDPOINT_URL="$MINIO_URL"

# Suppress the aws CLI pager.
export AWS_PAGER=""

# ── file naming ───────────────────────────────────────────────────────────────

TIMESTAMP="$(date -u '+%Y%m%dT%H%M%SZ')"
DUMP_FILENAME="${DB_NAME}_${TIMESTAMP}.sql.gz"
DUMP_KEY="${BACKUP_PREFIX}/${DUMP_FILENAME}"

# Use /tmp so we never need write access to the repo directory.
DUMP_TMPFILE="$(mktemp "/tmp/pg-backup-XXXXXX.sql.gz")"
trap 'rm -f "$DUMP_TMPFILE"' EXIT

# ── bucket creation (idempotent) ──────────────────────────────────────────────

log "Ensuring bucket '${BACKUP_BUCKET}' exists..."
aws s3 mb "s3://${BACKUP_BUCKET}" 2>/dev/null || true

# ── pg_dump ───────────────────────────────────────────────────────────────────

log "Starting pg_dump of '${DB_NAME}' on ${DB_HOST}:${DB_PORT}..."

# PGPASSWORD is the standard pg_dump password env var — unset by trap above.
export PGPASSWORD="$DB_PASSWORD"

# shellcheck disable=SC2086
if ! pg_dump \
      --host="$DB_HOST" \
      --port="$DB_PORT" \
      --username="$DB_USER" \
      --dbname="$DB_NAME" \
      --format=plain \
      --no-password \
      $PG_DUMP_EXTRA_OPTS \
    | gzip -9 > "$DUMP_TMPFILE"; then
  die 2 "pg_dump failed. Check DB connectivity and credentials."
fi

unset PGPASSWORD

DUMP_SIZE="$(du -h "$DUMP_TMPFILE" | cut -f1)"
log "Dump complete — compressed size: ${DUMP_SIZE}"

# ── dump sanity checks ────────────────────────────────────────────────────────
# Guard against pg_dump returning 0 for a logically-empty or truncated dump:
#   1. Minimum size: reject anything under 1 KB (a valid dump always includes
#      a header + schema even for an empty DB).
#   2. Footer marker: pg_dump always ends with "PostgreSQL database dump complete"
#      in the SQL output. Absence means the dump was cut short mid-stream.

DUMP_BYTES="$(wc -c < "$DUMP_TMPFILE")"
if [ "$DUMP_BYTES" -lt 1024 ]; then
  die 2 "Dump sanity check failed — compressed file is only ${DUMP_BYTES} bytes (< 1 KB). Aborting upload."
fi

log "Checking dump footer marker..."
if ! gzip -dc "$DUMP_TMPFILE" | grep -q "PostgreSQL database dump complete"; then
  die 2 "Dump sanity check failed — footer marker 'PostgreSQL database dump complete' not found. Dump may be truncated. Aborting upload."
fi
log "Dump sanity checks passed."

# ── upload to MinIO / S3 ──────────────────────────────────────────────────────

log "Uploading to s3://${BACKUP_BUCKET}/${DUMP_KEY} ..."

if ! aws s3 cp "$DUMP_TMPFILE" "s3://${BACKUP_BUCKET}/${DUMP_KEY}" \
    --storage-class STANDARD; then
  die 3 "Upload failed. Check MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, and bucket permissions."
fi

log "Upload complete: s3://${BACKUP_BUCKET}/${DUMP_KEY}"

# ── retention pruning ─────────────────────────────────────────────────────────
# List all dumps under the prefix, sort oldest first, delete anything beyond
# BACKUP_RETENTION_COUNT dumps. Non-fatal: a pruning failure should not cause
# the backup job to report failure — the backup itself succeeded.

log "Pruning dumps beyond ${BACKUP_RETENTION_COUNT} most recent under s3://${BACKUP_BUCKET}/${BACKUP_PREFIX}/..."

RETENTION_RC=0

# aws s3 ls output format: "2024-01-15 03:00:00    123456 filename"
DUMP_LIST="$(aws s3 ls "s3://${BACKUP_BUCKET}/${BACKUP_PREFIX}/" 2>/dev/null \
  | awk '{print $NF}' \
  | grep '\.sql\.gz$' \
  | sort)" || RETENTION_RC=$?

if [ "$RETENTION_RC" -ne 0 ]; then
  warn "Could not list dumps for retention pruning (exit ${RETENTION_RC}). Skipping."
else
  TOTAL="$(echo "$DUMP_LIST" | grep -c '.' || true)"
  KEEP="$BACKUP_RETENTION_COUNT"

  if [ "$TOTAL" -gt "$KEEP" ]; then
    TO_DELETE="$(echo "$DUMP_LIST" | head -n "$(( TOTAL - KEEP ))")"
    log "Deleting $(( TOTAL - KEEP )) old dump(s)..."
    while IFS= read -r fname; do
      [ -z "$fname" ] && continue
      log "  Deleting ${BACKUP_PREFIX}/${fname}"
      aws s3 rm "s3://${BACKUP_BUCKET}/${BACKUP_PREFIX}/${fname}" || {
        warn "Failed to delete ${fname} — continuing."
        RETENTION_RC=4
      }
    done <<< "$TO_DELETE"
  else
    log "Retention OK — ${TOTAL}/${KEEP} dumps stored, nothing to prune."
  fi
fi

# ── done ─────────────────────────────────────────────────────────────────────

log "Backup finished successfully: ${DUMP_FILENAME} (${DUMP_SIZE})"

if [ "$RETENTION_RC" -eq 4 ]; then
  warn "Some old dumps could not be pruned — check bucket permissions."
  exit 4
fi

exit 0
