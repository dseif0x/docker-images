FROM jellyfin/jellyfin:10.10.7

# Install common dependencies
RUN apt update && apt install -y wget gnupg2

# Install OpenCL dependencies
RUN mkdir --parents --mode=0755 /etc/apt/keyrings \
    && wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | \
    gpg --dearmor | tee /etc/apt/keyrings/rocm.gpg > /dev/null \
    && echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/rocm/apt/6.4.2 jammy main" \
    | tee /etc/apt/sources.list.d/rocm.list \
    && echo 'Package: *\nPin: release o=repo.radeon.com\nPin-Priority: 600' \
    | tee /etc/apt/preferences.d/rocm-pin-600 \
    && apt update

# Conditionally install rocm-opencl-runtime for amd64 architecture
RUN if [ "$(dpkg --print-architecture)" = "amd64" ]; then \
        apt install -y rocm-opencl-runtime; \
    fi
