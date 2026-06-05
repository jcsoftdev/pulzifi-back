#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${1:-.env}"
MODE="${2:-all}" # "docker" for make dev, "web" for make dev-web, "prod-docker" for make dev-with-prod, "prod-web" for make dev-web-with-prod, "all" for both

# Production DB port — used when connecting to the production database.
PROD_DB_PORT=5432

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

# Prod modes target the production DB: DB_PORT must stay 5432, not the local mapping.
case "$MODE" in
  prod-*) REGISTRY_PORTS[DB_PORT]="$PROD_DB_PORT" ;;
esac

# Which keys each mode needs
docker_keys=(HTTP_PORT SCRAPER_PORT DB_PORT DEV_REDIS_PORT LOCALSTACK_PORT GRPC_PORT)
web_keys=(DEV_WEB_PORT)

case "$MODE" in
  docker|prod-docker) active_keys=("${docker_keys[@]}") ;;
  web|prod-web)       active_keys=("${web_keys[@]}") ;;
  *)                  active_keys=(HTTP_PORT DEV_WEB_PORT SCRAPER_PORT DB_PORT DEV_REDIS_PORT LOCALSTACK_PORT GRPC_PORT) ;;
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

# Local dev runs over HTTPS (Go terminates TLS in-process, see cmd/server/main.go). HTTPS gives
# the browser a secure context on lvh.me so Chrome sends Sec-Fetch-* / cookies that Payload's
# admin auth needs — over plain HTTP the CMS admin loops on login.
upsert_env "NEXTJS_URL" "http://localhost:${web_port}"
upsert_env "FRONTEND_URL" "https://pulzifi.lvh.me:${api_port}"
upsert_env "NEXT_PUBLIC_SERVER_URL" "https://lvh.me:${api_port}"
upsert_env "SERVER_API_URL" "https://localhost:${api_port}"
upsert_env "OAUTH_REDIRECT_BASE_URL" "https://pulzifi.lvh.me:${api_port}"
upsert_env "INTEGRATION_OAUTH_REDIRECT_BASE" "https://pulzifi.lvh.me:${api_port}"
# Host-side Payload (s3Storage) talks to localstack on its mapped host port. The docker monolith
# overrides this with the internal `localstack:4566` via compose, so this value is host-only.
upsert_env "MINIO_ENDPOINT" "http://localhost:${localstack_port}"
upsert_env "MINIO_PUBLIC_URL" "http://localhost:${localstack_port}"
upsert_env "COOKIE_DOMAIN" ".lvh.me"
upsert_env "CORS_ALLOWED_ORIGINS" "https://localhost:${api_port},https://*.localhost:${api_port},http://localhost:${web_port},https://pulzifi.lvh.me:${api_port},https://*.pulzifi.lvh.me:${api_port},https://lvh.me:${api_port}"
upsert_env "PAYLOAD_CSRF_ORIGINS" "https://localhost:${api_port},https://pulzifi.lvh.me:${api_port},https://lvh.me:${api_port}"
# TLS cert paths are CONTAINER paths (repo mounted at /workspace). Unset in prod (Traefik does TLS).
upsert_env "TLS_CERT_FILE" "/workspace/tools/dev-certs/lvh.me.pem"
upsert_env "TLS_KEY_FILE" "/workspace/tools/dev-certs/lvh.me-key.pem"
# Host frontend SSR calls the HTTPS Go API — trust the mkcert root CA (host path, no spaces).
upsert_env "NODE_EXTRA_CA_CERTS" "${PWD}/tools/dev-certs/rootCA.pem"

if [ "$MODE" = "docker" ] || [ "$MODE" = "prod-docker" ]; then
  cat <<EOF

Dev ports (docker):
  API / Go entrypoint: http://localhost:${api_port}
  Scraper host port:  ${PORTS[SCRAPER_PORT]}
  Postgres host port: ${PORTS[DB_PORT]}
  Redis host port:    ${PORTS[DEV_REDIS_PORT]}
  LocalStack port:    ${PORTS[LOCALSTACK_PORT]}
  gRPC host port:     ${PORTS[GRPC_PORT]}
EOF
elif [ "$MODE" = "web" ] || [ "$MODE" = "prod-web" ]; then
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
