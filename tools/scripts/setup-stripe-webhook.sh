#!/usr/bin/env bash
# Create or update the Stripe webhook endpoint for Pulzifi (idempotent by URL).
#
# Use this ONLY for staging/production (publicly reachable URL).
# For local dev, run `stripe listen --forward-to localhost:3002/api/v1/billing/webhook` instead.
#
# Reads STRIPE_SECRET_KEY from .env. The webhook URL is the first positional arg or WEBHOOK_URL env.
#
# Usage:
#   ENV_FILE=.env.staging    ./tools/scripts/setup-stripe-webhook.sh https://staging.pulzifi.com/api/v1/billing/webhook
#   ENV_FILE=.env.production ./tools/scripts/setup-stripe-webhook.sh https://app.pulzifi.com/api/v1/billing/webhook
#
# After running, copy the printed `whsec_...` signing secret into the target env file as STRIPE_WEBHOOK_SECRET.

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

WEBHOOK_URL="${1:-${WEBHOOK_URL:-}}"
[ -z "$WEBHOOK_URL" ] && { err "Webhook URL required. Pass as arg or WEBHOOK_URL env."; exit 1; }
[[ "$WEBHOOK_URL" =~ ^https:// ]] || { err "Webhook URL must be HTTPS."; exit 1; }

# Events the Pulzifi billing handler listens to
EVENTS=(
  checkout.session.completed
  checkout.session.expired
  customer.created
  customer.updated
  customer.deleted
  customer.subscription.created
  customer.subscription.updated
  customer.subscription.deleted
  customer.subscription.paused
  customer.subscription.resumed
  customer.subscription.trial_will_end
  invoice.created
  invoice.finalized
  invoice.paid
  invoice.payment_failed
  invoice.payment_action_required
  invoice.upcoming
  payment_intent.succeeded
  payment_intent.payment_failed
  payment_method.attached
  payment_method.detached
)

info "Stripe mode: $STRIPE_MODE"
info "Webhook URL: $WEBHOOK_URL"
info ""

# Find existing endpoint with same URL
info "Looking up existing webhook endpoints..."
EXISTING=$(stripe_call webhook_endpoints list --limit 100 2>/dev/null)
EXISTING_ID=$(echo "$EXISTING" | jq -r --arg url "$WEBHOOK_URL" '.data[] | select(.url == $url) | .id' | head -n 1)

API_URL="https://api.stripe.com/v1"

build_payload() {
  echo "url=$WEBHOOK_URL"
  echo "description=Pulzifi billing — $STRIPE_MODE"
  for e in "${EVENTS[@]}"; do
    echo "enabled_events[]=$e"
  done
}

if [ -n "$EXISTING_ID" ]; then
  warn "Endpoint already exists: $EXISTING_ID — updating events list"
  # Update can't change URL, only events + description + status
  PAYLOAD=$(mktemp)
  {
    echo "description=Pulzifi billing — $STRIPE_MODE"
    for e in "${EVENTS[@]}"; do echo "enabled_events[]=$e"; done
  } > "$PAYLOAD"
  curl -s -u "$STRIPE_SECRET_KEY:" "$API_URL/webhook_endpoints/$EXISTING_ID" \
    --data-binary @"$PAYLOAD" > /tmp/webhook_result.json
  rm -f "$PAYLOAD"
  SECRET_HINT="Existing secret unchanged. Retrieve via: stripe webhook_endpoints retrieve $EXISTING_ID --api-key \$STRIPE_SECRET_KEY"
else
  info "Creating new webhook endpoint..."
  PAYLOAD=$(mktemp)
  build_payload > "$PAYLOAD"
  curl -s -u "$STRIPE_SECRET_KEY:" "$API_URL/webhook_endpoints" \
    --data-binary @"$PAYLOAD" > /tmp/webhook_result.json
  rm -f "$PAYLOAD"
  SECRET=$(jq -r '.secret // empty' /tmp/webhook_result.json)
  SECRET_HINT="STRIPE_WEBHOOK_SECRET=$SECRET"
fi

if jq -e '.error' /tmp/webhook_result.json >/dev/null 2>&1; then
  err "Webhook setup failed:"
  jq '.error' /tmp/webhook_result.json >&2
  exit 1
fi

ENDPOINT_ID=$(jq -r '.id' /tmp/webhook_result.json)
info ""
info "Done. Endpoint ID: $ENDPOINT_ID"
info ""
info "Add this to your $ENV_FILE:"
echo ""
echo "  $SECRET_HINT"
echo ""
warn "The secret is only shown ONCE for new endpoints. Save it now."
