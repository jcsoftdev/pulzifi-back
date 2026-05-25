#!/usr/bin/env bash
# Shared helper for all Stripe automation scripts.
#
# Resolution order for STRIPE_SECRET_KEY (first non-empty wins):
#   1. Already-exported shell env (cloud injection: Dokploy, direnv, dotenvx, k8s, etc.)
#   2. Local .env file (loaded via `set -a; source` — respects shell expansion)
#   3. Override via ENV_FILE=path/to/file
#
# This means local dev uses .env, and prod uses whatever your cloud platform injects —
# without ever needing a .env.production file.
#
# Usage from another script:
#   source "$(dirname "$0")/stripe-env.sh"
#   stripe_call prices list --limit 10

set -euo pipefail

# 1. If STRIPE_SECRET_KEY already in shell (cloud env or already-sourced .env) → use it as-is
# 2. Otherwise try to load from local .env
if [ -z "${STRIPE_SECRET_KEY:-}" ]; then
  ENV_FILE="${ENV_FILE:-.env}"
  if [ -f "$ENV_FILE" ]; then
    # Proper sourcing — respects quotes, expansions, multi-line. set -a exports every assignment.
    set -a
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    set +a
  fi
fi

if [ -z "${STRIPE_SECRET_KEY:-}" ]; then
  cat >&2 <<EOF
ERROR: STRIPE_SECRET_KEY not found.

Resolution order:
  1. Export in shell: export STRIPE_SECRET_KEY=sk_test_...
  2. Local .env file (current path: ${ENV_FILE:-.env})
  3. ENV_FILE=path/to/file ./script.sh

For Dokploy/cloud prod: vars should already be injected. SSH into the container
and run the script directly — it picks up STRIPE_SECRET_KEY from the environment.
EOF
  exit 1
fi

# Detect mode from key prefix
if [[ "$STRIPE_SECRET_KEY" == sk_live_* ]]; then
  STRIPE_MODE="LIVE"
elif [[ "$STRIPE_SECRET_KEY" == sk_test_* ]]; then
  STRIPE_MODE="TEST"
else
  echo "ERROR: STRIPE_SECRET_KEY has unexpected prefix (expected sk_test_ or sk_live_)" >&2
  exit 1
fi

# Confirmation gate for live mode (prevent accidents on cloud-injected prod keys)
if [ "$STRIPE_MODE" = "LIVE" ] && [ "${STRIPE_LIVE_CONFIRMED:-}" != "yes" ]; then
  cat >&2 <<EOF

⚠️  LIVE MODE detected (sk_live_ key).
   This will modify your REAL Stripe account: real customers, real products, real money.

EOF
  read -r -p "   Type 'I UNDERSTAND' to continue: " confirm
  [ "$confirm" = "I UNDERSTAND" ] || { echo "Aborted." >&2; exit 0; }
  export STRIPE_LIVE_CONFIRMED=yes
fi

export STRIPE_SECRET_KEY
export STRIPE_MODE

# Wrapper that injects --api-key into every stripe CLI call.
stripe_call() {
  stripe "$@" --api-key "$STRIPE_SECRET_KEY"
}
export -f stripe_call
