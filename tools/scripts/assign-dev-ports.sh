#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${1:-.env}"

# Local development ports exposed on the host. Container-internal ports stay fixed.
declare -A DEFAULT_PORTS=(
  [HTTP_PORT]=3002
  [DEV_WEB_PORT]=3003
  [SCRAPER_PORT]=3005
  [DB_PORT]=5436
  [DEV_REDIS_PORT]=6382
  [LOCALSTACK_PORT]=4567
  [GRPC_PORT]=9000
)

ordered_keys=(HTTP_PORT DEV_WEB_PORT SCRAPER_PORT DB_PORT DEV_REDIS_PORT LOCALSTACK_PORT GRPC_PORT)

if [ ! -f "$ENV_FILE" ]; then
  if [ -f ".env.example" ]; then
    cp .env.example "$ENV_FILE"
    echo "Created $ENV_FILE from .env.example"
  else
    touch "$ENV_FILE"
  fi
fi

get_env_value() {
  local key="$1"
  local value=""
  if [ -f "$ENV_FILE" ]; then
    value=$(grep -E "^${key}=" "$ENV_FILE" | tail -n 1 | cut -d= -f2- || true)
  fi
  printf '%s' "$value"
}

is_port_busy() {
  local port="$1"
  lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
}

contains_port() {
  local needle="$1"
  shift || true
  local port
  for port in "$@"; do
    [ "$port" = "$needle" ] && return 0
  done
  return 1
}

pick_port() {
  local requested="$1"
  shift || true
  local reserved=("$@")
  local candidate="$requested"

  while is_port_busy "$candidate" || contains_port "$candidate" "${reserved[@]}"; do
    candidate=$((candidate + 1))
  done

  printf '%s' "$candidate"
}

upsert_env() {
  local key="$1"
  local value="$2"

  if grep -qE "^${key}=" "$ENV_FILE"; then
    tmp=$(mktemp)
    awk -v key="$key" -v value="$value" 'BEGIN{done=0} $0 ~ "^" key "=" { if (!done) { print key "=" value; done=1 } next } { print }' "$ENV_FILE" > "$tmp"
    mv "$tmp" "$ENV_FILE"
  else
    printf '\n%s=%s\n' "$key" "$value" >> "$ENV_FILE"
  fi
}

selected=()
declare -A PORTS

for key in "${ordered_keys[@]}"; do
  current="$(get_env_value "$key")"
  if ! [[ "$current" =~ ^[0-9]+$ ]]; then
    current="${DEFAULT_PORTS[$key]}"
  fi

  chosen="$(pick_port "$current" "${selected[@]}")"
  PORTS[$key]="$chosen"
  selected+=("$chosen")

  if [ "$chosen" != "$current" ]; then
    echo "$key=$current is busy; using $chosen"
  else
    echo "$key=$chosen"
  fi

  upsert_env "$key" "$chosen"
done

api_port="${PORTS[HTTP_PORT]}"
web_port="${PORTS[DEV_WEB_PORT]}"
localstack_port="${PORTS[LOCALSTACK_PORT]}"

upsert_env "NEXTJS_URL" "http://localhost:${web_port}"
upsert_env "FRONTEND_URL" "http://pulzifi.local:${api_port}"
upsert_env "NEXT_PUBLIC_SERVER_URL" "http://localhost:${api_port}"
upsert_env "OAUTH_REDIRECT_BASE_URL" "http://localhost:${api_port}"
upsert_env "INTEGRATION_OAUTH_REDIRECT_BASE" "http://localhost:${api_port}"
upsert_env "MINIO_PUBLIC_URL" "http://localhost:${localstack_port}"
upsert_env "CORS_ALLOWED_ORIGINS" "http://localhost:${api_port},http://*.localhost:${api_port},http://localhost:${web_port},http://pulzifi.local:${api_port},http://*.pulzifi.local:${api_port}"

cat <<EOF

Dev ports ready:
  API / Go entrypoint: http://localhost:${api_port}
  Next.js dev server: http://localhost:${web_port}
  Scraper host port:  ${PORTS[SCRAPER_PORT]}
  Postgres host port: ${PORTS[DB_PORT]}
  Redis host port:    ${PORTS[DEV_REDIS_PORT]}
  LocalStack port:    ${PORTS[LOCALSTACK_PORT]}
  gRPC host port:     ${PORTS[GRPC_PORT]}
EOF
