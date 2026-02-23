# ============================================================
# Dockerfile - Frontend (Next.js bun workspace monorepo)
# Build context: repo root (Railway resolves dockerfilePath from repo root)
# ============================================================

FROM oven/bun:1 AS builder

WORKDIR /app

# Copy workspace root manifests first for layer caching
COPY frontend/package.json frontend/bun.lockb frontend/turbo.json ./

# Copy all package manifests before install
COPY frontend/apps/web/package.json ./apps/web/
COPY frontend/packages/ui/package.json ./packages/ui/
COPY frontend/packages/services/package.json ./packages/services/
COPY frontend/packages/shared-http/package.json ./packages/shared-http/
COPY frontend/packages/notix/package.json ./packages/notix/
COPY frontend/packages/typescript-config/package.json ./packages/typescript-config/

RUN bun install --frozen-lockfile

# Copy source
COPY frontend/apps/ ./apps/
COPY frontend/packages/ ./packages/

ENV NODE_ENV=production

RUN bun run build

# ============================================================
# Runtime
# ============================================================
FROM oven/bun:1

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000

# Copy build output and runtime dependencies
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/packages ./packages
COPY --from=builder /app/apps/web/.next ./apps/web/.next
COPY --from=builder /app/apps/web/public ./apps/web/public
COPY --from=builder /app/apps/web/package.json ./apps/web/package.json
COPY --from=builder /app/apps/web/next.config.mjs ./apps/web/next.config.mjs

EXPOSE 3000

WORKDIR /app/apps/web

CMD ["bun", "run", "start"]
