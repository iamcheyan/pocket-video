# Roadmap

## Implemented

### Application

- React/TypeScript PocketJS surface with six screens: home, favorites, folder
  view, library, video/audio detail, queue, settings.
- Go REST service with SQLite, ordered migrations, and a single source of truth
  for settings, library content, tasks, and media paths.
- Deterministic mock mode: four folders, twelve videos, real MP4/M4A downloads,
  extracted cover art, and an offline player — no Bilibili account required.
- Search by title and creator, favorite-folder filtering, deep-linkable routes
  (`#detail/<id>`, `#favorites/<id>`).
- Download tasks with progress, pause/resume from the persisted byte offset,
  duplicate suppression, failed-state errors, and queue cleanup.
- Manual and automatic favorite sync on a configurable interval.
- Settings for language, theme, cache directory, download directory, sync
  interval, auto-sync, and proxy, validated on the server.
- English, Simplified Chinese, and Japanese with compile-time and runtime key
  parity; system/light/dark themes that follow the device while in system mode.
- Responsive layouts from 320×240 to 1920×1080, compact handheld list mode,
  verified 1×–3× DPI rendering, keyboard/gamepad/touch/mouse input.
- 21 verified screenshots and `make test` covering Go, the frontend, the
  production build, image integrity, and horizontal-overflow checks.

### Live provider

- yt-dlp executes only inside Go, with `--continue`, proxy support, progress
  parsing, and clear task errors when the binary or network is unavailable.

## Next

1. **Bilibili login / cookie import.** A supported credential flow (QR login or
   user-supplied cookie file outside the repository) that feeds
   `YtDlpProvider.SyncFavorites`. Until then, sync returns
   `provider_unavailable` instead of inventing data.
2. **BBDownProvider.** Implement `VideoProvider` with BBDown for higher-quality
   Bilibili streams and multi-part videos; no API or UI changes required.
3. **Cover art for live downloads.** Reuse `generateDemoPoster` after a yt-dlp
   run, or consume yt-dlp thumbnails.
4. **Library management.** Delete a downloaded item, free-space reporting, and a
   "keep only N recent" retention rule owned by the service.
5. **Queue concurrency.** A configurable worker limit instead of one worker per
   request, plus a persisted retry policy with backoff.
6. **Playback memory.** Store last position per video and offer "continue" on
   the detail screen.
7. **Accessibility polish.** Announced route changes, reduced-motion handling,
   and a focus-order audit at 320×240 with a real gamepad.

## PJOS / PocketJS Runtime milestones

1. Replace `fetch` in `api.ts` with the runtime's IPC bridge; keep JSON payloads.
2. Map `/api/media` and `/api/cover` to runtime media/file handles.
3. Define the `.pj` manifest and package the static build plus the Go binary
   (see [PACKAGING.md](PACKAGING.md)).
4. Validate codecs, storage permissions, and background sync on device.
5. Publish PocketVideo as the reference app for providers, SQLite, and tasks.

## Not implemented

- Authenticated Bilibili favorite extraction and login. No account or cookies
  were available; the live provider reports an explicit error.
- Real Bilibili media in the mock catalog. Demo clips are generated to stay
  deterministic and redistributable; their length is not the catalog duration.
- A `.pj` package or a PJOS runtime build; the migration path is documented.
- Multi-user or remote deployments. The service is a localhost application.

## Known runtime limits

- Browsers play only the codecs they support. Demo video is H.264/AAC and audio
  is AAC, both playable in current Chrome/Chromium.
- Media seeking/resume depends on HTTP Range support, which `http.ServeFile`
  provides; a future runtime must offer an equivalent handle.
- The 320×240 layout is usable but dense by design; long titles truncate in the
  compact list rather than wrapping the row.
