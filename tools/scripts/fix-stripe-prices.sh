#!/usr/bin/env bash
# Backfill lookup_keys and fix tax_behavior on EXISTING Stripe prices.
#
# Run this once after creating prices manually in the Dashboard so they have stable lookup_keys
# (which the other scripts rely on).
#
# Usage:
#   ./tools/scripts/fix-stripe-prices.sh                          # test
#   ENV_FILE=.env.production ./tools/scripts/fix-stripe-prices.sh # live

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/stripe-env.sh"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info() { echo -e "${GREEN}$*${NC}"; }
warn() { echo -e "${YELLOW}$*${NC}"; }
err()  { echo -e "${RED}$*${NC}" >&2; }

command -v jq >/dev/null || { err "jq required"; exit 1; }

info "Stripe mode: $STRIPE_MODE"
PRICES=$(stripe_call prices list --limit 100 --expand 'data.product' 2>/dev/null)

declare -a MAPPINGS=(
  "starter|month|starter_monthly"
  "starter|year|starter_yearly"
  "pro|month|pro_monthly"
  "pro|year|pro_yearly"
)

for mapping in "${MAPPINGS[@]}"; do
  IFS='|' read -r name_match interval target_key <<< "$mapping"

  price=$(echo "$PRICES" | jq -r --arg name "$name_match" --arg interval "$interval" '
    .data[]
    | select(.active == true)
    | select(.product.name | ascii_downcase | contains($name | ascii_downcase))
    | select(.recurring.interval == $interval)
    | {id, lookup_key, tax_behavior, product_name: .product.name}' | jq -s '.[0]')

  if [ "$price" = "null" ] || [ -z "$price" ]; then
    warn "  ⚠ no price found for: $name_match $interval"
    continue
  fi

  price_id=$(echo "$price" | jq -r '.id')
  current_key=$(echo "$price" | jq -r '.lookup_key // empty')
  current_tax=$(echo "$price" | jq -r '.tax_behavior // empty')
  product_name=$(echo "$price" | jq -r '.product_name')

  info ""
  info "=== $product_name ($interval) — $price_id ==="

  if [ "$current_key" = "$target_key" ]; then
    warn "  ↻ lookup_key already: $target_key"
  elif [ -n "$current_key" ]; then
    warn "  ↻ existing lookup_key '$current_key' kept (use --transfer-lookup-key manually if you want to change)"
  else
    info "  + setting lookup_key: $target_key"
    stripe_call prices update "$price_id" --lookup-key "$target_key" --transfer-lookup-key >/dev/null 2>&1 \
      && info "    ✓ ok" \
      || err "    ✗ failed"
  fi

  if [ "$current_tax" = "exclusive" ] || [ "$current_tax" = "inclusive" ]; then
    warn "  ↻ tax_behavior already: $current_tax"
  else
    info "  + setting tax_behavior: exclusive"
    stripe_call prices update "$price_id" -d "tax_behavior=exclusive" >/dev/null 2>&1 \
      && info "    ✓ ok" \
      || err "    ✗ failed (tax_behavior is immutable once set — recreate the price)"
  fi
done

info ""
info "Done."
