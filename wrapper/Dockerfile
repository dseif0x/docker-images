# Builds WorldObservationLog/wrapper, a tool to decrypt Apple Music songs.
#
# Upstream keeps amd64 and arm64 on separate branches, each with its own
# Android-target toolchain and its own bundled libraries:
#   - linux/amd64 -> "main"  branch  (x86_64-linux-android22)
#   - linux/arm64 -> "arm64" branch  (aarch64-linux-android22)
# Both cross-compile with the linux-x86_64 Android NDK prebuilt, so the build
# stage always runs on the native builder ($BUILDPLATFORM) and only the runtime
# stage matches the requested target platform.
ARG NDK_VERSION=23

FROM --platform=$BUILDPLATFORM debian:13.2 AS build
ARG NDK_VERSION
ARG TARGETARCH
SHELL ["/bin/bash", "-c"]
WORKDIR /app

RUN --mount=type=cache,target=/var/lib/apt,sharing=locked \
    --mount=type=cache,target=/var/cache/apt,sharing=locked \
    apt-get update && \
    apt-get install -y \
        build-essential cmake unzip git wget ca-certificates lsb-release gnupg aria2 && \
    if [ "$TARGETARCH" = "arm64" ]; then \
        apt-get install -y gcc-aarch64-linux-gnu; \
    fi

# Clone the upstream branch matching the target architecture.
RUN case "$TARGETARCH" in \
        amd64) BRANCH=main ;; \
        arm64) BRANCH=arm64 ;; \
        *) echo "Unsupported TARGETARCH: $TARGETARCH" >&2; exit 1 ;; \
    esac && \
    git clone --depth 1 --branch "$BRANCH" \
        https://github.com/WorldObservationLog/wrapper.git /app

# Host clang/LLVM toolchain used during the build.
RUN bash -c "$(wget -O - https://apt.llvm.org/llvm.sh)"

# Android NDK r23b (linux-x86_64 prebuilt) — CMakeLists.txt expects it next to
# the sources and cross-compiles to this arch's Android target.
RUN aria2c -o android-ndk.zip \
        "https://dl.google.com/android/repository/android-ndk-r${NDK_VERSION}b-linux.zip" && \
    unzip -q -d /app android-ndk.zip && \
    rm android-ndk.zip

RUN mkdir -p build && \
    cmake -S /app -B /app/build && \
    cmake --build /app/build -j"$(nproc)"

FROM debian:13.2

WORKDIR /app
COPY --from=build /app/wrapper /app/wrapper
COPY --from=build /app/rootfs /app/rootfs
COPY --from=build /app/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/wrapper

# 10020 decrypt, 20020 M3U8, 30020 account.
EXPOSE 10020 20020 30020

CMD ["/app/entrypoint.sh"]
