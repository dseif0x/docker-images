# wrapper

Container image for [WorldObservationLog/wrapper](https://github.com/WorldObservationLog/wrapper),
a tool to decrypt Apple Music songs. An active Apple Music subscription is
required.

The image is assembled from upstream's official **prebuilt release artifacts**
rather than compiled from source. Upstream's from-source build pins Dobby to a
moving `master` that currently fails to link, breaking their own CI; the
prebuilt path (which their own Dockerfile supports) avoids that and also skips
the heavy Android NDK build. Both `linux/amd64` and `linux/arm64` are supported
— upstream publishes a separate per-arch release, and the Dockerfile downloads
the one matching the build's target architecture.

The release tags are overridable via build args (`WRAPPER_AMD64_TAG`,
`WRAPPER_AMD64_ASSET`, `WRAPPER_ARM64_TAG`, `WRAPPER_ARM64_ASSET`) if you need
to pin a specific upstream release.

## Ports

- `10020` — decryption service
- `20020` — M3U8 playlist service
- `30020` — account service

## Usage

The container needs `--privileged` and a persistent volume for
`/app/rootfs/data` (stores authentication and config).

On first run the account database is missing, so login credentials must be
provided via the `USERNAME` and `PASSWORD` environment variables:

```bash
docker run --privileged \
  -v ./rootfs/data:/app/rootfs/data \
  -p 10020:10020 -p 20020:20020 -p 30020:30020 \
  -e USERNAME="your-apple-id" \
  -e PASSWORD="your-password" \
  ghcr.io/dseif0x/wrapper:latest
```

Once the account database exists in the mounted volume, the credentials are no
longer needed:

```bash
docker run --privileged \
  -v ./rootfs/data:/app/rootfs/data \
  -p 10020:10020 -p 20020:20020 -p 30020:30020 \
  ghcr.io/dseif0x/wrapper:latest
```

Extra wrapper arguments can be appended after the image name and are forwarded
to the `wrapper` binary by the entrypoint.
