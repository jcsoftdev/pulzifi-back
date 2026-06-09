#!/usr/bin/env bash
# Create or update Pulzifi products + prices in Stripe (idempotent via lookup_keys).
#
# Reads STRIPE_SECRET_KEY from .env. Same script works in test and live mode.
#
# Usage:
#   ./tools/scripts/setup-stripe-products.sh                          # test (.env)
#   ENV_FILE=.env.production ./tools/scripts/setup-stripe-products.sh # live mode (requires "I UNDERSTAND" confirm)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/stripe-env.sh"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info() { echo -e "${GREEN}$*${NC}"; }
warn() { echo -e "${YELLOW}$*${NC}"; }
err()  { echo -e "${RED}$*${NC}" >&2; }

command -v jq >/dev/null || { err "jq required"; exit 1; }

# Plan catalog — single source of truth
# Format: code|product_name|description|monthly_cents|yearly_cents|statement_descriptor
read -r -d '' PLANS_TABLE <<'EOF' || true
starter|Pulzifi Starter|Perfect for solopreneurs, individual users and business owners. 1 workspace, up to 5 pages, 1 user account, 4 AI insights, 1 week storage, email + messages alerts.|2700|26400|PULZIFI STARTER
pro|Pulzifi Professional|Perfect for growing businesses ready to scale. Unlimited workspaces, up to 25 pages, 5 users, unlimited AI insights, multi-channel alerts (Email, Messages, Teams, Slack, Telegram), 1 month storage, priority support.|5400|55200|PULZIFI PRO
EOF

TAX_CODE="txcd_10000000"

info "Stripe mode: $STRIPE_MODE"

find_product_id() {
  local name="$1"
  stripe_call products list --limit 100 2>/dev/null \
    | jq -r --arg name "$name" '.data[] | select(.active and ((.name | ascii_downcase) == ($name | ascii_downcase))) | .id' \
    | head -n 1
}

find_price_by_lookup() {
  local key="$1"
  stripe_call prices list --limit 100 --lookup-keys "$key" 2>/dev/null \
    | jq -r '.data[] | select(.active) | .id' \
    | head -n 1
}

create_or_get_product() {
  local name="$1" description="$2" statement_descriptor="$3"
  local existing
  existing=$(find_product_id "$name")
  if [ -n "$existing" ]; then
    warn "  ↻ product exists: $name ($existing)" >&2
    echo "$existing"
    return
  fi
  info "  + creating product: $name" >&2
  stripe_call products create \
    --name "$name" \
    --description "$description" \
    --tax-code "$TAX_CODE" \
    -d "statement_descriptor=$statement_descriptor" \
    2>/dev/null | jq -r '.id'
}

create_or_get_price() {
  local product_id="$1" lookup_key="$2" amount_cents="$3" interval="$4"
  local existing
  existing=$(find_price_by_lookup "$lookup_key")
  if [ -n "$existing" ]; then
    warn "  ↻ price exists: $lookup_key ($existing)" >&2
    echo "$existing"
    return
  fi
  info "  + creating price: $lookup_key ($amount_cents cents / $interval)" >&2
  stripe_call prices create \
    --product "$product_id" \
    --unit-amount "$amount_cents" \
    --currency usd \
    --lookup-key "$lookup_key" \
    -d "recurring[interval]=$interval" \
    -d "tax_behavior=exclusive" \
    2>/dev/null | jq -r '.id'
}

while IFS='|' read -r code name description monthly_cents yearly_cents statement_descriptor; do
  [ -z "$code" ] && continue
  info ""
  info "=== Plan: $code ==="
  product_id=$(create_or_get_product "$name" "$description" "$statement_descriptor")
  create_or_get_price "$product_id" "${code}_monthly" "$monthly_cents" "month" >/dev/null
  create_or_get_price "$product_id" "${code}_yearly" "$yearly_cents" "year" >/dev/null
done <<< "$PLANS_TABLE"

info ""
info "Done. Next: ./tools/scripts/sync-stripe-prices.sh"
