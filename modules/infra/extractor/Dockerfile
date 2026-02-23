FROM oven/bun:1

WORKDIR /app

# Install basic system dependencies
RUN apt-get update && apt-get install -y \
    wget \
    gnupg \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY modules/infra/extractor/package.json .
COPY modules/infra/extractor/bun.lockb .

# Install node dependencies
RUN bun install

# Install playwright system dependencies
RUN ./node_modules/.bin/playwright install-deps

# Install chromium browser binary
RUN ./node_modules/.bin/playwright install chromium

COPY modules/infra/extractor/index.ts .

EXPOSE 3000

CMD ["bun", "run", "index.ts"]
