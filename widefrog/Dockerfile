FROM alpine:3.20 AS n_m3u8dl
WORKDIR /app
ARG TARGETARCH
ENV VERSION=v0.3.0-beta
ENV FILENAME_BASE=N_m3u8DL-RE_${VERSION}_linux
RUN apk add --no-cache curl tar 
# Determine correct filename based on architecture
RUN ARCHIVE_NAME="$FILENAME_BASE-$TARGETARCH.tar.gz" && \
    case "$TARGETARCH" in \
      amd64) ARCHIVE_NAME="${FILENAME_BASE}-x64_20241203.tar.gz" ;; \
      arm64) ARCHIVE_NAME="${FILENAME_BASE}-arm64_20241203.tar.gz" ;; \
      *) echo "Unsupported architecture: $TARGETARCH" && exit 1 ;; \
    esac && \
    echo "Downloading https://github.com/nilaoda/N_m3u8DL-RE/releases/download/$VERSION/$ARCHIVE_NAME" && \
    curl -L -o tool.tar.gz https://github.com/nilaoda/N_m3u8DL-RE/releases/download/$VERSION/$ARCHIVE_NAME && \
    tar -xzf tool.tar.gz


FROM debian:bookworm-slim AS bento4
WORKDIR /app
RUN apt update && apt install -y git cmake build-essential
RUN git clone https://github.com/axiomatic-systems/Bento4.git
WORKDIR /app/Bento4/cmakebuild
RUN cmake -DCMAKE_BUILD_TYPE=Release ..
RUN cmake --build . --target mp4decrypt


FROM alpine:3.20 AS shaka-packager
WORKDIR /app
ARG TARGETARCH
ENV VERSION=v3.4.2
ENV FILENAME_BASE=packager-linux
RUN apk add --no-cache curl tar 
RUN ARCHIVE_NAME="$FILENAME_BASE-$TARGETARCH.tar.gz" && \
    case "$TARGETARCH" in \
      amd64) ARCHIVE_NAME="${FILENAME_BASE}-x64" ;; \
      arm64) ARCHIVE_NAME="${FILENAME_BASE}-arm64" ;; \
      *) echo "Unsupported architecture: $TARGETARCH" && exit 1 ;; \
    esac && \
    echo "Downloading https://github.com/shaka-project/shaka-packager/releases/download/$VERSION/$ARCHIVE_NAME" && \
    curl -L -o shaka-packager https://github.com/shaka-project/shaka-packager/releases/download/$VERSION/$ARCHIVE_NAME


FROM alpine:3.20 AS widefrog
RUN apk add --no-cache curl unzip git
RUN curl -L -o widefrog.zip https://files.videohelp.com/u/308780/widefrog_v3_2_0_python_source_code.zip
RUN unzip widefrog.zip -d /app
COPY service.patch /app/service.patch
WORKDIR /app
RUN git apply --reject --whitespace=fix service.patch


# --- Stage 1: Build dependencies ---
FROM python:3.10-slim-bookworm AS builder
WORKDIR /app

# Install build tools only here
RUN apt update && apt install -y build-essential

# Copy and install Python deps
COPY --from=widefrog /app/requirements.txt /app
RUN pip install --upgrade pip \
 && grep -v '^#' requirements.txt | xargs -n 1 pip install

# Upgrade urllib3 and install Playwright
RUN pip install urllib3 --upgrade
RUN playwright install


# --- Stage 2: Final image ---
FROM python:3.10-slim-bookworm
WORKDIR /app

# Only install runtime dependencies
RUN apt update && apt install -y ffmpeg mkvtoolnix && apt clean && rm -rf /var/lib/apt/lists/*

# Copy installed packages from builder
COPY --from=builder /usr/local/lib/python3.10 /usr/local/lib/python3.10
COPY --from=builder /usr/local/bin /usr/local/bin
COPY --from=builder /usr/local/include /usr/local/include
COPY --from=builder /usr/local/share /usr/local/share
COPY --from=widefrog /app/ /app

# Make scripts executable
RUN chmod +x terminal/*

# Copy binary tools from other images
COPY --from=n_m3u8dl /app/N_m3u8DL-RE /usr/local/bin/N_m3u8DL-RE
COPY --from=bento4 /app/Bento4/cmakebuild/mp4decrypt /usr/local/bin/mp4decrypt
COPY --from=shaka-packager /app/shaka-packager /usr/local/bin/shaka-packager