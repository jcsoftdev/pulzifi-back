#!/usr/bin/env bash
# Configure the Stripe Customer Portal via API (idempotent — updates existing default config).
#
# Reads STRIPE_SECRET_KEY + STRIPE_PORTAL_RETURN_URL from shell env or sourced .env.
#
# Usage:
#   ./tools/scripts/setup-stripe-portal.sh                          # test
#   ENV_FILE=.env.production ./tools/scripts/setup-stripe-portal.sh # live

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/stripe-env.sh"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info() { echo -e "${GREEN}$*${NC}"; }
warn() { echo -e "${YELLOW}$*${NC}"; }
err()  { echo -e "${RED}$*${NC}" >&2; }

command -v jq >/dev/null || { err "jq required"; exit 1; }
command -v curl >/dev/null || { err "curl required"; exit 1; }

RETURN_URL="${STRIPE_PORTAL_RETURN_URL:-}"
[ -z "$RETURN_URL" ] && { err "STRIPE_PORTAL_RETURN_URL not set in shell env or sourced .env"; exit 1; }

PRIVACY_URL="${STRIPE_PRIVACY_URL:-https://pulzifi.com/privacy}"
TERMS_URL="${STRIPE_TERMS_URL:-https://pulzifi.com/terms}"
HEADLINE="${STRIPE_PORTAL_HEADLINE:-Manage your Pulzifi subscription}"

info "Stripe mode: $STRIPE_MODE"
info "Return URL:  $RETURN_URL"
info ""
info "Resolving Pulzifi prices via lookup_keys (starter_*, pro_*)..."

# Use lookup_keys as the canonical identifier — name matching is fragile when
# stale duplicate products exist in Stripe (e.g. "Starter monthly" leftovers).
PRICES_JSON=$(stripe_call prices list --limit 100 --expand 'data.product' \
  --lookup-keys starter_monthly --lookup-keys starter_yearly \
  --lookup-keys pro_monthly --lookup-keys pro_yearly 2>/dev/null)

resolve_price()   { echo "$PRICES_JSON" | jq -r --arg k "$1" '.data[] | select(.active and .lookup_key == $k) | .id'      | head -n 1; }
resolve_product() { echo "$PRICES_JSON" | jq -r --arg k "$1" '.data[] | select(.active and .lookup_key == $k) | .product.id' | head -n 1; }

# Each entry: "<lookup_key>". The portal config builds one product-entry per
# price because the current Stripe state has a separate product per interval
# (legacy from earlier setup runs). Stripe rejects a portal product entry whose
# prices don't all belong to the same product, so we enumerate per-interval.
PORTAL_KEYS=(starter_monthly starter_yearly pro_monthly pro_yearly)

declare -a PORTAL_PRODUCT_IDS=()
declare -a PORTAL_PRICE_IDS=()

for key in "${PORTAL_KEYS[@]}"; do
  price_id=$(resolve_price   "$key")
  prod_id=$( resolve_product "$key")
  if [ -z "$price_id" ] || [ -z "$prod_id" ]; then
    warn "  ↷ skipping $key (no active price)"
    continue
  fi
  PORTAL_PRICE_IDS+=("$price_id")
  PORTAL_PRODUCT_IDS+=("$prod_id")
  info "  ✓ $key → product=$prod_id price=$price_id"
done

[ "${#PORTAL_PRICE_IDS[@]}" -eq 0 ] && { err "No active prices found for any of: ${PORTAL_KEYS[*]}"; err "Run ./tools/scripts/setup-stripe-products.sh first."; exit 1; }

API_URL="https://api.stripe.com/v1"

info ""
info "Looking up existing default portal configuration..."
EXISTING_CONFIG=$(curl -s -u "$STRIPE_SECRET_KEY:" "$API_URL/billing_portal/configurations?is_default=true&limit=1")
CONFIG_ID=$(echo "$EXISTING_CONFIG" | jq -r '.data[0].id // empty')

# Build curl --data-urlencode args (curl handles encoding + & separators)
build_curl_args() {
  local args=(
    --data-urlencode "business_profile[headline]=$HEADLINE"
    --data-urlencode "business_profile[privacy_policy_url]=$PRIVACY_URL"
    --data-urlencode "business_profile[terms_of_service_url]=$TERMS_URL"
    --data-urlencode "default_return_url=$RETURN_URL"
    --data-urlencode "features[customer_update][enabled]=true"
    --data-urlencode "features[customer_update][allowed_updates][]=email"
    --data-urlencode "features[customer_update][allowed_updates][]=address"
    --data-urlencode "features[customer_update][allowed_updates][]=tax_id"
    --data-urlencode "features[invoice_history][enabled]=true"
    --data-urlencode "features[payment_method_update][enabled]=true"
    --data-urlencode "features[subscription_cancel][enabled]=true"
    --data-urlencode "features[subscription_cancel][mode]=at_period_end"
    --data-urlencode "features[subscription_cancel][cancellation_reason][enabled]=true"
    --data-urlencode "features[subscription_cancel][cancellation_reason][options][]=too_expensive"
    --data-urlencode "features[subscription_cancel][cancellation_reason][options][]=missing_features"
    --data-urlencode "features[subscription_cancel][cancellation_reason][options][]=switched_service"
    --data-urlencode "features[subscription_cancel][cancellation_reason][options][]=customer_service"
    --data-urlencode "features[subscription_cancel][cancellation_reason][options][]=low_quality"
    --data-urlencode "features[subscription_cancel][cancellation_reason][options][]=other"
    --data-urlencode "features[subscription_update][enabled]=true"
    --data-urlencode "features[subscription_update][default_allowed_updates][]=price"
    --data-urlencode "features[subscription_update][default_allowed_updates][]=promotion_code"
    --data-urlencode "features[subscription_update][proration_behavior]=create_prorations"
  )
  local i
  for i in "${!PORTAL_PRODUCT_IDS[@]}"; do
    args+=(--data-urlencode "features[subscription_update][products][${i}][product]=${PORTAL_PRODUCT_IDS[$i]}")
    args+=(--data-urlencode "features[subscription_update][products][${i}][prices][]=${PORTAL_PRICE_IDS[$i]}")
  done
  printf '%s\n' "${args[@]}"
}

# Collect args into array
mapfile -t CURL_ARGS < <(build_curl_args)

if [ -n "$CONFIG_ID" ]; then
  info "Updating existing config: $CONFIG_ID"
  curl -s -u "$STRIPE_SECRET_KEY:" "$API_URL/billing_portal/configurations/$CONFIG_ID" "${CURL_ARGS[@]}" > /tmp/portal_result.json
else
  info "Creating new default config..."
  curl -s -u "$STRIPE_SECRET_KEY:" "$API_URL/billing_portal/configurations" "${CURL_ARGS[@]}" > /tmp/portal_result.json
fi

if jq -e '.error' /tmp/portal_result.json >/dev/null 2>&1; then
  err "Portal configuration failed:"
  jq '.error' /tmp/portal_result.json >&2
  exit 1
fi

CONFIG_ID=$(jq -r '.id' /tmp/portal_result.json)
info ""
info "Done. Portal configuration ID: $CONFIG_ID"
info "Default return URL:           $(jq -r '.default_return_url' /tmp/portal_result.json)"
