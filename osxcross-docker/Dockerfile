FROM ubuntu:24.04 
RUN apt update \
&& apt install -y  \
    ca-certificates  \
    gpg \
    git \
    wget \
    software-properties-common \
    lsb-release \
    ninja-build \
    clang lld \
    pkg-config \
    cmake \
    libxml2-dev \
    libssl-dev \
    bash \
    patch \
    make \
    tar \
    xz-utils \
    bzip2 \
    gzip \
    sed \
    cpio \
    libbz2-dev \
    zlib1g-dev \
    libplist-dev \
    libplist++-dev

# Setup OSXCross for macOS builds
WORKDIR /opt
RUN git clone https://github.com/tpoechtrager/osxcross.git

WORKDIR /opt/osxcross
ARG MACOS_SDK_URL
RUN wget -P tarballs/ "https://github.com/joseluisq/macosx-sdks/releases/download/26.1/MacOSX26.1.sdk.tar.xz"

RUN UNATTENDED=1 ENABLE_ARCHS="arm64 x86_64" OSX_VERSION_MIN=26.0 ./build.sh
RUN ENABLE_COMPILER_RT_INSTALL=1 ./build_compiler_rt.sh
ENV PATH="/opt/osxcross/target/bin:${PATH}"
ENV OSXCROSS_MP_INC=1
RUN cat <<EOF1 >> /opt/osxcross/tools/toolchain.cmake
set(CMAKE_OBJC_COMPILER "\${OSXCROSS_TARGET_DIR}/bin/\${OSXCROSS_HOST}-clang")
set(CMAKE_OBJCXX_COMPILER "\${OSXCROSS_TARGET_DIR}/bin/\${OSXCROSS_HOST}-clang++")

set(CMAKE_STRIP "\${OSXCROSS_TARGET_DIR}/bin/\${OSXCROSS_HOST}-strip")
EOF1

#RUN ls -l /opt/osxcross/target/bin/ && exit 1

RUN git clone https://github.com/ProcursusTeam/ldid.git /tmp/ldid && \
    cd /tmp/ldid && \
    make CC=clang CXX=clang++ INSTALLPREFIX=/usr/local install && \
    cd / && rm -rf /tmp/ldid && \
    ldid --version || echo "ldid installed"