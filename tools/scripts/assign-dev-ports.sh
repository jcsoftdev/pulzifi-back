#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${1:-.env}"
MODE="${2:-all}" # "docker" for make dev, "web" for make dev-web, "all" for both

# Ports from port-registry MCP — source of truth.
declare -A REGISTRY_PORTS=(
  [HTTP_PORT]=3002
  [DEV_WEB_PORT]=3003
  [SCRAPER_PORT]=3005
  [DB_PORT]=5436
  [DEV_REDIS_PORT]=6382
  [LOCALSTACK_PORT]=4567
  [GRPC_PORT]=9000
)

# Which keys each mode needs
docker_keys=(HTTP_PORT SCRAPER_PORT DB_PORT DEV_REDIS_PORT LOCALSTACK_PORT GRPC_PORT)
web_keys=(DEV_WEB_PORT)

case "$MODE" in
  docker) active_keys=("${docker_keys[@]}") ;;
  web)    active_keys=("${web_keys[@]}") ;;
  *)      active_keys=(HTTP_PORT DEV_WEB_PORT SCRAPER_PORT DB_PORT DEV_REDIS_PORT LOCALSTACK_PORT GRPC_PORT) ;;
esac

if [ ! -f "$ENV_FILE" ]; then
  if [ -f ".env.example" ]; then
    cp .env.example "$ENV_FILE"
    echo "Created $ENV_FILE from .env.example"
  else
    touch "$ENV_FILE"
  fi
fi

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

get_env_value() {
  local key="$1"
  if [ -f "$ENV_FILE" ]; then
    grep -E "^${key}=" "$ENV_FILE" | tail -n 1 | cut -d= -f2- || true
  fi
}

declare -A PORTS

# Always ensure all ports are in .env (docker-compose reads DEV_WEB_PORT too)
for key in HTTP_PORT DEV_WEB_PORT SCRAPER_PORT DB_PORT DEV_REDIS_PORT LOCALSTACK_PORT GRPC_PORT; do
  port="${REGISTRY_PORTS[$key]}"
  PORTS[$key]="$port"
  upsert_env "$key" "$port"
done

# Print only the active keys
for key in "${active_keys[@]}"; do
  echo "$key=${PORTS[$key]}"
done

api_port="${PORTS[HTTP_PORT]}"
web_port="${PORTS[DEV_WEB_PORT]}"
localstack_port="${PORTS[LOCALSTACK_PORT]}"

upsert_env "NEXTJS_URL" "http://localhost:${web_port}"
upsert_env "FRONTEND_URL" "http://pulzifi.lvh.me:${api_port}"
upsert_env "NEXT_PUBLIC_SERVER_URL" "http://pulzifi.lvh.me:${api_port}"
upsert_env "SERVER_API_URL" "http://localhost:${api_port}"
upsert_env "OAUTH_REDIRECT_BASE_URL" "http://pulzifi.lvh.me:${api_port}"
upsert_env "INTEGRATION_OAUTH_REDIRECT_BASE" "http://pulzifi.lvh.me:${api_port}"
upsert_env "MINIO_PUBLIC_URL" "http://localhost:${localstack_port}"
upsert_env "COOKIE_DOMAIN" ".lvh.me"
upsert_env "CORS_ALLOWED_ORIGINS" "http://localhost:${api_port},http://*.localhost:${api_port},http://localhost:${web_port},http://pulzifi.lvh.me:${api_port},http://*.pulzifi.lvh.me:${api_port},http://lvh.me:${api_port}"
upsert_env "PAYLOAD_CSRF_ORIGINS" "http://localhost:${api_port},http://localhost:${web_port},http://pulzifi.lvh.me:${api_port}"

if [ "$MODE" = "docker" ]; then
  cat <<EOF

Dev ports (docker):
  API / Go entrypoint: http://localhost:${api_port}
  Scraper host port:  ${PORTS[SCRAPER_PORT]}
  Postgres host port: ${PORTS[DB_PORT]}
  Redis host port:    ${PORTS[DEV_REDIS_PORT]}
  LocalStack port:    ${PORTS[LOCALSTACK_PORT]}
  gRPC host port:     ${PORTS[GRPC_PORT]}
EOF
elif [ "$MODE" = "web" ]; then
  cat <<EOF

Dev ports (web):
  Next.js dev server: http://localhost:${web_port}
EOF
else
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
fi
