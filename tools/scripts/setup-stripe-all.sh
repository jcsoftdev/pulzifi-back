#!/usr/bin/env bash
# One-shot Stripe environment setup orchestrator.
#
# Runs all idempotent setup scripts in order:
#   1. setup-stripe-products.sh — products + prices with lookup_keys
#   2. fix-stripe-prices.sh     — backfills lookup_keys/tax_behavior on manually-created prices
#   3. sync-stripe-prices.sh    — pushes price IDs into the plans table
#   4. setup-stripe-portal.sh   — Customer Portal configuration
#   5. setup-stripe-webhook.sh  — (optional) production webhook endpoint
#
# Usage:
#   ./tools/scripts/setup-stripe-all.sh                                       # test mode (.env)
#   ENV_FILE=.env.production ./tools/scripts/setup-stripe-all.sh https://pulzifi.com/api/v1/billing/webhook
#   ./tools/scripts/setup-stripe-all.sh --skip-webhook                        # everything except webhook (e.g. local dev)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info() { echo -e "${GREEN}$*${NC}"; }
warn() { echo -e "${YELLOW}$*${NC}"; }

SKIP_WEBHOOK=false
WEBHOOK_URL=""
for arg in "$@"; do
  case "$arg" in
    --skip-webhook) SKIP_WEBHOOK=true ;;
    https://*) WEBHOOK_URL="$arg" ;;
  esac
done

info "════════════════════════════════════════"
info "1/5 — Creating products + prices"
info "════════════════════════════════════════"
"$SCRIPT_DIR/setup-stripe-products.sh"

info ""
info "════════════════════════════════════════"
info "2/5 — Backfilling lookup_keys + tax_behavior on existing prices"
info "════════════════════════════════════════"
"$SCRIPT_DIR/fix-stripe-prices.sh"

info ""
info "════════════════════════════════════════"
info "3/5 — Syncing price IDs to database"
info "════════════════════════════════════════"
"$SCRIPT_DIR/sync-stripe-prices.sh"

info ""
info "════════════════════════════════════════"
info "4/5 — Configuring Customer Portal"
info "════════════════════════════════════════"
"$SCRIPT_DIR/setup-stripe-portal.sh"

if [ "$SKIP_WEBHOOK" = "true" ]; then
  warn ""
  warn "5/5 — Skipped webhook endpoint (--skip-webhook)"
  warn "       For local dev: stripe listen --forward-to localhost:3002/api/v1/billing/webhook"
elif [ -n "$WEBHOOK_URL" ]; then
  info ""
  info "════════════════════════════════════════"
  info "5/5 — Creating webhook endpoint"
  info "════════════════════════════════════════"
  "$SCRIPT_DIR/setup-stripe-webhook.sh" "$WEBHOOK_URL"
else
  warn ""
  warn "5/5 — Webhook URL not provided. Skipping."
  warn "       For prod, run:"
  warn "         ./tools/scripts/setup-stripe-webhook.sh https://pulzifi.com/api/v1/billing/webhook"
  warn "       For local dev, run:"
  warn "         stripe listen --forward-to localhost:3002/api/v1/billing/webhook"
fi

info ""
info "════════════════════════════════════════"
info "All done."
info "════════════════════════════════════════"
