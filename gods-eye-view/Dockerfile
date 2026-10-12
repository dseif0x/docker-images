# God's Eye View (https://github.com/bilawalsidhu/gods-eye-view), MIT.
#
# Built once with `vite build`, served with `vite preview`: the API proxies
# (flights, satellites, CCTV, ...) are Vite middleware and run under preview
# as well as dev. Bump GEV_REF to update.
#
# Keys: none are needed to start. Browser-exposed keys (GOOGLE_MAPS_API_KEY,
# CESIUM_ION_TOKEN, MAPILLARY_CLIENT_TOKEN) are inlined at BUILD time, so this
# keyless image starts on Esri/OSM imagery. Server-side keys (OPENSKY_*,
# AISSTREAM_API_KEY, FIRMS_MAP_KEY, TOMTOM_API_KEY, OPENAI_API_KEY, ...) are
# read at runtime from the environment. The in-app POWER UP panel only
# accepts loopback clients, so behind a reverse proxy keys come from env.
ARG NODE_IMAGE=node:24-bookworm-slim

FROM alpine:3.22 AS source
ARG GEV_REF=591f299d11f38a612629a274463196d57ae3862e
RUN apk add --no-cache git \
 && git init -q /src \
 && git -C /src fetch -q --depth 1 https://github.com/bilawalsidhu/gods-eye-view.git "$GEV_REF" \
 && git -C /src checkout -q FETCH_HEAD \
 && rm -rf /src/.git

FROM ${NODE_IMAGE} AS build
WORKDIR /app
# Puppeteer is a dev dependency for QA scripts only; skip its Chrome download.
ENV PUPPETEER_SKIP_DOWNLOAD=1
COPY --from=source /src/package.json /src/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY --from=source /src/ ./
RUN npm run build && rm -rf node_modules/.vite

FROM ${NODE_IMAGE}
ENV NODE_ENV=production \
    PUPPETEER_SKIP_DOWNLOAD=1 \
    HOST=0.0.0.0 \
    PORT=4173 \
    OPENSKY_AUTH_MODE=anon
WORKDIR /app
# vite, its plugins and the server code are needed at runtime (preview
# loads vite.config.js), so the full tree is kept. The server writes
# .gev-cache and .gev-logs into /app: owned by the runtime user (node, 1000).
COPY --from=build --chown=1000:1000 /app /app
USER 1000:1000
EXPOSE 4173
CMD ["sh", "-c", "exec node node_modules/vite/bin/vite.js preview --host \"$HOST\" --port \"$PORT\" --strictPort"]
