# DEVICE VERIFICATION — pocket-video

Status: **DEVICE-PASS** for the captured states. Local host verification is in
`apps/pocket-video/POCKETJS-VERIFICATION.md`.

## Device

- Anbernic RG34XX, Base OS 1.5.2 (h700), kernel 4.9.170 aarch64.
- Access: Debian host -> HX90 -> USB ADB.
- Display: `/dev/fb0`, 720x480 visible page, 32 bpp.
- Runtime: `/mnt/sdcard/System/pocketjs/bin/pocket-linux-kms` (existing boot runtime, not rebuilt or replaced).
- Package: `pocket-video` sha256 `7548161fe8ce3148414b8d3d24932bf6c7eb26fe9a48fc726373f6dd4d727849`.
- Deployed to `/storage/pjs/pocket-video/app.pocket` only.

## Procedure

1. Recorded boot launcher pid `21225` and `boot.pocket`
   sha256 `6ed41612bce1723d1eda64b5a0ee9333a105ed498b7085c03267539cf6c1b32d`.
2. Paused the launcher supervisor, stopped the launcher, and started the
   package with the same runtime and `/dev/fb0`.
3. Created a virtual keyboard with `/dev/uinput` and drove the native view
   tree through a FIFO (`uinputd-c`); no host browser and no HTML.
4. Captured `/dev/fb0` with `dd` -> base64 -> PIL.
5. Killed the package, restored the launcher, and read the pid
   `16016` and boot hash back.

Result: launcher pid changed, boot package sha256 stayed
`6ed41612bce1723d1eda64b5a0ee9333a105ed498b7085c03267539cf6c1b32d`. No write to `/storage/System`, `/storage/pocketjs`,
`/storage/baseos.conf` or `/storage/content`.

## Frames

The images are the device framebuffer. The UI was rasterized by the PocketJS
native renderer on the RG34XX, not by host Chrome.

| file | sha256 (first 16) |
|---|---|
| `docs/device/device-video-detail-action.png` | `9f213b09385db995` |
| `docs/device/device-video-detail.png` | `a83a7790882689e5` |
| `docs/device/device-video-fb-after.png` | `d59130e3329b8496` |
| `docs/device/device-video-fb-before.png` | `d59130e3329b8496` |
| `docs/device/device-video-home-focus.png` | `7a07e0a2a8fe1b20` |
| `docs/device/device-video-home.png` | `5b8c2ac1d356cc08` |
| `docs/device/device-video-settings-toggled.png` | `62b3e8c71e3ffaf3` |
| `docs/device/device-video-settings.png` | `3556abebd9431a80` |

`device-video-fb-before.png` and `device-video-fb-after.png` are the boot
launcher before and after; they are byte-identical here, which is the intended
restore evidence.

Captured states: 6 states: home, focus, detail, detail action, settings, settings toggle.
