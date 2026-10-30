# Packaging for PJOS / PocketJS Runtime

This document records the intended migration from "desktop Linux app" to a
first-class PocketJS reference app. Nothing here is implemented yet; the current
code is shaped so the migration is packaging work, not a rewrite.

## What already helps

- `fetch` is confined to `frontend/src/api.ts`; pages import the typed client.
- Media and artwork are reached through URLs, not filesystem paths
  (`/api/media/{id}`, `/api/cover/{id}`).
- The service owns SQLite, downloads, and yt-dlp; the UI has no Node, Electron,
  or shell access.
- `GET /api/account` already exposes mode/provider/auth state, which a runtime
  permission model can fill in.
- The frontend is a static Vite build with no server-side rendering.

## Proposed `.pj` layout

```text
pocketvideo.pj/
  manifest.json
  web/                     static Vite output (index.html + assets)
  bin/pocketvideo-service  Go binary for the target architecture
  data/                    created at runtime under the app data directory
  README.md
```

`manifest.json` (draft):

```json
{
  "id": "com.pocketvideo.app",
  "name": "PocketVideo",
  "version": "1.0.0",
  "runtime": "pocketjs",
  "entry": "web/index.html",
  "services": [
    { "name": "api", "exec": "bin/pocketvideo-service", "protocol": "http", "socket": "api.sock" }
  ],
  "permissions": ["storage:app-data", "network:bilibili", "media:playback", "process:yt-dlp"],
  "input": ["touch", "keyboard", "gamepad"],
  "viewports": { "min": [320, 240], "max": [1920, 1080] }
}
```

## Runtime mapping

| Current desktop behavior | PJOS/PocketJS target |
| --- | --- |
| `fetch('/api/...')` over localhost | runtime IPC/HTTP bridge to the same JSON contract |
| `http.ServeFile` media URL | runtime media handle or packaged streaming API |
| `POCKETVIDEO_DATA` directory | app-data sandbox path from the manifest |
| yt-dlp subprocess | declared `process` permission or a bundled downloader capability |
| Automatic sync goroutine | background service slot with battery/network policy |
| `prefers-color-scheme` | runtime theme signal, same `theme.ts` entry point |
| JSON locale files in the bundle | runtime locale negotiation feeding `detectLocale` |

## Migration checklist

1. **Bundle.** `npm run build` plus a cross-compiled Go binary for the target
   architecture; write `manifest.json` and validate it with the runtime tooling.
2. **Transport.** Add a PocketJS adapter implementing the `ApiClient` shape in
   `api.ts`; keep every payload identical. No page changes should be required.
3. **Media.** Map `/api/media` and `/api/cover` to runtime handles; verify Range
   seeking and pause/resume still work.
4. **Storage.** Move SQLite, cache, and downloads under the sandboxed app-data
   directory; ensure the service creates them with restrictive permissions.
5. **Permissions.** Request network access for Bilibili, storage for downloads,
   and process permission for yt-dlp; degrade to mock mode when denied.
6. **Device QA.** Re-run the 320×240 / 480×320 / 720×720 captures on hardware,
   verify gamepad mapping, codecs, and thermal/battery behavior of automatic sync.
7. **Login.** Implement the runtime's secure credential store for Bilibili; never
   write cookies into the package or repository.

## Non-goals

- Shipping Bilibili content or credentials inside the package.
- Requiring a desktop compositor, Electron, or a Node runtime.
- Changing the REST contract per platform; the contract is the portability layer.
