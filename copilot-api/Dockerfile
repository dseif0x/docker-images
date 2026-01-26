FROM alpine:latest AS source_dl
WORKDIR /app
RUN apk add --no-cache git
RUN git clone https://github.com/ericc-ch/copilot-api.git 
WORKDIR /app/copilot-api

FROM scratch AS source
COPY --from=source_dl /app/copilot-api /

FROM oven/bun:1.2.19-alpine AS builder
WORKDIR /app

COPY --from=source ./package.json ./bun.lock ./
RUN bun install --frozen-lockfile

COPY --from=source . .
RUN bun run build

FROM oven/bun:1.2.19-alpine AS runner
WORKDIR /app

COPY --from=source ./package.json ./bun.lock ./
RUN bun install --frozen-lockfile --production --ignore-scripts --no-cache

COPY --from=builder /app/dist ./dist

EXPOSE 4141

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --spider -q http://localhost:4141/ || exit 1

COPY --from=source entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]