#!/usr/bin/env bash
# Sync Stripe price IDs → public.plans table.
#
# Reads STRIPE_SECRET_KEY + DB_* from shell env or sourced .env.
# Same script targets local (DB_HOST=localhost) or remote (DB_HOST=cloud-host) — whichever
# the env points to. No docker exec — uses psql directly so it works for cloud DBs.
#
# Usage:
#   ./tools/scripts/sync-stripe-prices.sh                          # uses .env (whichever DB it points to)
#   ENV_FILE=.env.production ./tools/scripts/sync-stripe-prices.sh # different env file
#   DATABASE_URL=postgres://... ./tools/scripts/sync-stripe-prices.sh   # override connection directly

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/stripe-env.sh"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info() { echo -e "${GREEN}$*${NC}"; }
warn() { echo -e "${YELLOW}$*${NC}"; }
err()  { echo -e "${RED}$*${NC}" >&2; }

command -v jq >/dev/null   || { err "jq required (brew install jq)"; exit 1; }
command -v psql >/dev/null || { err "psql required (brew install libpq && brew link --force libpq)"; exit 1; }

# Build DATABASE_URL from individual DB_* vars (already loaded by stripe-env.sh) unless explicit.
if [ -z "${DATABASE_URL:-}" ]; then
  [ -z "${DB_HOST:-}" ]     && { err "DB_HOST not set";     exit 1; }
  [ -z "${DB_PORT:-}" ]     && { err "DB_PORT not set";     exit 1; }
  [ -z "${DB_NAME:-}" ]     && { err "DB_NAME not set";     exit 1; }
  [ -z "${DB_USER:-}" ]     && { err "DB_USER not set";     exit 1; }
  [ -z "${DB_PASSWORD:-}" ] && { err "DB_PASSWORD not set"; exit 1; }
  DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
  DB_DISPLAY="${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
else
  DB_DISPLAY="(DATABASE_URL override)"
fi

info "Stripe mode: $STRIPE_MODE"
info "Target DB:   $DB_DISPLAY"
info ""
info "Fetching prices from Stripe..."
PRICES_JSON=$(stripe_call prices list --limit 100 --expand 'data.product' 2>/dev/null)

find_price() {
  local lookup_key="$1" name_match="$2" interval="$3"
  local found
  found=$(echo "$PRICES_JSON" | jq -r --arg key "$lookup_key" '
    .data[] | select(.active == true) | select(.lookup_key == $key) | .id' | head -n 1)
  if [ -n "$found" ]; then echo "$found"; return; fi
  echo "$PRICES_JSON" | jq -r --arg name "$name_match" --arg interval "$interval" '
    .data[]
    | select(.active == true)
    | select(.product.name | ascii_downcase | contains($name | ascii_downcase))
    | select(.recurring.interval == $interval)
    | .id' | head -n 1
}

STARTER_MONTHLY=$(find_price "starter_monthly" "starter" "month")
STARTER_YEARLY=$(find_price "starter_yearly" "starter" "year")
PRO_MONTHLY=$(find_price "pro_monthly" "pro" "month")
PRO_YEARLY=$(find_price "pro_yearly" "pro" "year")

missing=0
[ -z "$STARTER_MONTHLY" ] && { err "Missing: starter_monthly"; missing=1; }
[ -z "$STARTER_YEARLY" ]  && { err "Missing: starter_yearly"; missing=1; }
[ -z "$PRO_MONTHLY" ]     && { err "Missing: pro_monthly"; missing=1; }
[ -z "$PRO_YEARLY" ]      && { err "Missing: pro_yearly"; missing=1; }
[ $missing -eq 1 ] && { err "Run ./tools/scripts/setup-stripe-products.sh first."; exit 1; }

cat <<EOF

Matched prices ($STRIPE_MODE):
  starter_monthly: $STARTER_MONTHLY
  starter_yearly:  $STARTER_YEARLY
  pro_monthly:     $PRO_MONTHLY
  pro_yearly:      $PRO_YEARLY

EOF

if [ "${RUN_NONINTERACTIVE:-}" != "yes" ]; then
  read -r -p "Apply to $DB_DISPLAY? [y/N] " confirm
  [[ "$confirm" =~ ^[yY]$ ]] || { warn "Aborted."; exit 0; }
fi

SQL=$(cat <<SQL
UPDATE public.plans
SET stripe_price_id_monthly = '${STARTER_MONTHLY}',
    stripe_price_id_yearly  = '${STARTER_YEARLY}'
WHERE code = 'starter';

UPDATE public.plans
SET stripe_price_id_monthly = '${PRO_MONTHLY}',
    stripe_price_id_yearly  = '${PRO_YEARLY}'
WHERE code = 'pro';

SELECT code, stripe_price_id_monthly, stripe_price_id_yearly
FROM public.plans
ORDER BY code;
SQL
)

info "Updating plans table..."
echo "$SQL" | psql "$DATABASE_URL"
info "Done."
