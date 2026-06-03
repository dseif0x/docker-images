# wrapper

Container image for [WorldObservationLog/wrapper](https://github.com/WorldObservationLog/wrapper),
a tool to decrypt Apple Music songs. An active Apple Music subscription is
required.

The image clones and builds the upstream sources with the Android NDK r23b
toolchain. Both `linux/amd64` and `linux/arm64` are supported: upstream keeps
the two architectures on separate branches (`main` for x86_64, `arm64` for
aarch64), each with its own Android-target toolchain and bundled libraries, so
the Dockerfile clones the branch matching the build's target architecture.

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
