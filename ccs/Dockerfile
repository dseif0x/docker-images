FROM alpine:latest AS source_dl
WORKDIR /ccs

RUN apk add --no-cache git

RUN git clone https://github.com/kaitranntt/ccs.git /ccs --branch=v7.56.0 --depth=1

# =============================================================================
# Build stage: compile TypeScript and build UI
# =============================================================================
FROM node:20-bookworm-slim AS build

SHELL ["/bin/bash", "-lc"]

# Pin bun version for reproducible builds
ARG BUN_VERSION=1.2.21
ENV BUN_INSTALL=/usr/local/bun
ENV PATH="$BUN_INSTALL/bin:$PATH"

RUN apt-get update \
  && apt-get install -y --no-install-recommends curl ca-certificates unzip \
  && rm -rf /var/lib/apt/lists/*

# Install specific bun version
RUN curl -fsSL https://bun.sh/install | bash -s "bun-v${BUN_VERSION}"

WORKDIR /app

# Dependency install layer (avoid running install scripts inside the image build)
COPY --from=source_dl /ccs/package.json /ccs/bun.lock /ccs/bunfig.toml ./
COPY --from=source_dl /ccs/ui/package.json /ccs/ui/bun.lock ./ui/

RUN bun install --frozen-lockfile --ignore-scripts \
  && (cd ui && bun install --frozen-lockfile --ignore-scripts)

COPY --from=source_dl /ccs .

RUN bun run build:all

# Validate build artifacts exist
RUN test -d dist && test -d lib && echo "[OK] Build artifacts validated"

# =============================================================================
# Runtime stage: minimal production image
# =============================================================================
FROM node:20-bookworm-slim AS runtime

SHELL ["/bin/bash", "-lc"]

# Pin bun version for reproducible builds
ARG BUN_VERSION=1.2.21
ENV BUN_INSTALL=/usr/local/bun
ENV PATH="$BUN_INSTALL/bin:/home/node/.opencode/bin:$PATH"

RUN apt-get update \
  && apt-get install -y --no-install-recommends curl ca-certificates unzip \
  && rm -rf /var/lib/apt/lists/*

# Install specific bun version
RUN curl -fsSL https://bun.sh/install | bash -s "bun-v${BUN_VERSION}"

WORKDIR /app

COPY --from=source_dl /ccs/package.json /ccs/bun.lock ./
RUN bun install --frozen-lockfile --production --ignore-scripts

COPY --from=source_dl /ccs/docker/entrypoint.sh /usr/local/bin/ccs-entrypoint
RUN chmod +x /usr/local/bin/ccs-entrypoint

COPY --from=build /app/dist ./dist
COPY --from=build /app/lib ./lib
COPY --from=build /app/config ./config
COPY --from=build /app/scripts ./scripts
COPY --from=build /app/README.md ./README.md
COPY --from=build /app/LICENSE ./LICENSE

# Install AI CLI tools (using latest - pin versions in production if needed)
# These are optional tools for docker exec usage
RUN npm install -g @google/gemini-cli @vibe-kit/grok-cli @anthropic-ai/claude-code \
  && npm install -g @kaitranntt/ccs --force \
  && npm cache clean --force \
  && su -s /bin/bash node -c 'curl -fsSL https://opencode.ai/install | bash -s -- --no-modify-path' \
  && ln -sf /app/dist/ccs.js /usr/local/bin/ccs

ENV CCS_PORT=3000
EXPOSE 3000 8317

ENTRYPOINT ["/usr/local/bin/ccs-entrypoint"]

CMD ["bash", "-c", "node dist/ccs.js config --port ${CCS_PORT}"]