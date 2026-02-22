# ============================================================
# Dockerfile - Frontend (Next.js bun workspace monorepo)
# Build context: frontend/ (set Railway root dir to "frontend")
# ============================================================

FROM oven/bun:1 AS builder

WORKDIR /app

# Copy workspace root manifests first for layer caching
COPY package.json bun.lockb turbo.json ./

# Copy all package manifests before install
COPY apps/web/package.json ./apps/web/
COPY packages/ui/package.json ./packages/ui/
COPY packages/services/package.json ./packages/services/
COPY packages/shared-http/package.json ./packages/shared-http/
COPY packages/notix/package.json ./packages/notix/
COPY packages/typescript-config/package.json ./packages/typescript-config/

RUN bun install --frozen-lockfile

# Copy source
COPY apps/ ./apps/
COPY packages/ ./packages/

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
