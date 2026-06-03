# Packages WorldObservationLog/wrapper, a tool to decrypt Apple Music songs.
#
# Upstream's from-source build pins Dobby to a moving `master` that currently
# fails to link (undefined __android_log_vprint), which breaks upstream's own
# CI for both architectures. Instead we use upstream's official prebuilt
# release artifacts — the same "prebuild" path their own Dockerfile supports.
# amd64 and arm64 ship as separate per-arch releases, each bundling the wrapper
# launcher plus the proprietary rootfs libraries (which cannot be built from
# source in any case).
#
# Override these ARGs to pin a different upstream release.
ARG WRAPPER_AMD64_TAG=wrapper.x86_64.latest
ARG WRAPPER_AMD64_ASSET=Wrapper.x86_64.latest.zip
ARG WRAPPER_ARM64_TAG=wrapper.arm64.latest
ARG WRAPPER_ARM64_ASSET=Wrapper.arm64.latest.zip

# Fetch runs on the native builder (no compilation, just download + unzip); the
# downloaded artifact is selected by the target architecture.
FROM --platform=$BUILDPLATFORM debian:13.2 AS fetch
ARG TARGETARCH
ARG WRAPPER_AMD64_TAG WRAPPER_AMD64_ASSET WRAPPER_ARM64_TAG WRAPPER_ARM64_ASSET
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl unzip && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app
RUN set -eux; \
    case "$TARGETARCH" in \
        amd64) TAG="$WRAPPER_AMD64_TAG"; ASSET="$WRAPPER_AMD64_ASSET" ;; \
        arm64) TAG="$WRAPPER_ARM64_TAG"; ASSET="$WRAPPER_ARM64_ASSET" ;; \
        *) echo "Unsupported TARGETARCH: $TARGETARCH" >&2; exit 1 ;; \
    esac; \
    curl -fsSL -o /tmp/wrapper.zip \
        "https://github.com/WorldObservationLog/wrapper/releases/download/${TAG}/${ASSET}"; \
    unzip -q /tmp/wrapper.zip -d /app; \
    rm /tmp/wrapper.zip; \
    test -f /app/wrapper && test -d /app/rootfs && test -f /app/entrypoint.sh

FROM debian:13.2

WORKDIR /app
COPY --from=fetch /app/wrapper /app/wrapper
COPY --from=fetch /app/rootfs /app/rootfs
COPY --from=fetch /app/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/wrapper /app/entrypoint.sh

# 10020 decrypt, 20020 M3U8, 30020 account.
EXPOSE 10020 20020 30020

CMD ["/app/entrypoint.sh"]
