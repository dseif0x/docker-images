FROM alpine:3.21 AS downloader
ARG TARGETARCH

RUN apk add --no-cache curl tar

RUN set -eux; \
    case "${TARGETARCH}" in \
        amd64) \
            CCACHE_ARCH="x86_64-musl-static"; \
            STORAGE_ARCH="amd64"; \
            ;; \
        arm64) \
            CCACHE_ARCH="aarch64-musl-static"; \
            STORAGE_ARCH="arm64"; \
            ;; \
        *) echo "Unsupported architecture: ${TARGETARCH}" && exit 1 ;; \
    esac; \
    \
    curl -fsSL "https://github.com/ccache/ccache/releases/download/v4.13.6/ccache-4.13.6-linux-${CCACHE_ARCH}.tar.gz" \
        | tar -xz --strip-components=1 -C /usr/local/bin/ --wildcards '*/ccache'; \
    \
    curl -fsSL "https://github.com/ccache/ccache-storage-http-go/releases/download/v0.7/ccache-storage-http-go-0.7-linux-${STORAGE_ARCH}.tar.gz" \
        | tar -xz --strip-components=1 -C /usr/local/bin/ --wildcards '*/ccache-storage-http'; \
    \
    chmod +x /usr/local/bin/ccache /usr/local/bin/ccache-storage-http


FROM scratch

COPY --from=downloader /usr/local/bin/ccache /ccache
COPY --from=downloader /usr/local/bin/ccache-storage-http /ccache-storage-https