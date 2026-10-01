# Rockchip device plugin

Kubernetes device plugin for Rockchip SoCs (RK3588 etc.). Replaces the separate
`vpu-device-plugin` and the upstream `npu-device-plugin` with one binary that
advertises two resources:

| Resource | Devices |
|---|---|
| `rockchip.com/npu` | DRM nodes bound to the `rknpu` driver (looked up via sysfs, not hardcoded to `renderD129`), `/dev/rknpu` or `/dev/accel/*` (mainline `rocket`); plus a read-only mount of `/proc/device-tree/compatible` for rknn-toolkit |
| `rockchip.com/vpu` | `/dev/mpp_service` (required), `/dev/rga`, `/dev/mali0`, `/dev/dma_heap/*` and all other `/dev/dri` nodes, **excluding** the NPU's |

A resource is only advertised if its devices exist. Each is offered as
`--npu-replicas` / `--vpu-replicas` slots (default 1) that all map to the same
hardware, so several pods can share it. Device nodes are re-checked every 30s
and reported unhealthy if they disappear.

`device-plugin -list` prints what would be advertised and exits.

The plugin pod needs the host's `/dev`, `/sys` and `/proc` and the kubelet
device-plugin directory mounted; see the `rockchip-device-plugin` chart in
[helm-charts](https://github.com/dseif0x/helm-charts).
