FROM oven/bun:1-debian AS base

# Install Chromium + dependencies + fonts
RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    fonts-liberation \
    fonts-noto-color-emoji \
    fonts-noto-cjk \
    libnss3 \
    libatk-bridge2.0-0 \
    libdrm2 \
    libxkbcommon0 \
    libgbm1 \
    libasound2 \
    libatspi2.0-0 \
    libxshmfence1 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxrandr2 \
    libpango-1.0-0 \
    libcairo2 \
    libcups2 \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install dependencies
COPY modules/infra/scraper/package.json modules/infra/scraper/bun.lock* ./
RUN bun install --frozen-lockfile || bun install

# Copy source
COPY modules/infra/scraper/ .

# Create non-root user
RUN groupadd -r scraper && useradd -r -g scraper -G audio,video scraper \
    && mkdir -p /home/scraper && chown -R scraper:scraper /home/scraper /app

USER scraper

# Use system-installed Chromium — patchright patches at CDP level, not binary
ENV CHROMIUM_PATH=/usr/bin/chromium
ENV PORT=3000
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD curl -sf http://localhost:3000/health || exit 1

CMD ["bun", "run", "index.ts"]
