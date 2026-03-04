FROM oven/bun:1-debian

WORKDIR /app

# Install Chromium system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget gnupg curl \
    libasound2 libatk-bridge2.0-0 libatk1.0-0 libcairo2 libcups2 \
    libdbus-1-3 libdrm2 libexpat1 libgbm1 libglib2.0-0 libnspr4 \
    libnss3 libpango-1.0-0 libx11-6 libx11-xcb1 libxcb1 libxcomposite1 \
    libxdamage1 libxext6 libxfixes3 libxrandr2 libxshmfence1 \
    libxkbcommon0 \
    fonts-liberation fonts-noto-color-emoji xdg-utils \
    && rm -rf /var/lib/apt/lists/*

COPY modules/infra/extractor/package.json .
COPY modules/infra/extractor/bun.lockb .

# Install node dependencies
RUN bun install

# Download Chromium binary using playwright-core CLI (rebrowser-playwright uses same binaries)
RUN ./node_modules/.bin/playwright-core install chromium

COPY modules/infra/extractor/index.ts .

EXPOSE 3000

CMD ["bun", "run", "index.ts"]
