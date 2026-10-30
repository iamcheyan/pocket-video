# Architecture

PocketVideo keeps the browser as a thin TypeScript/React client. It calls a Go HTTP service using JSON. The service is the only business and persistence boundary: it owns SQLite (`modernc.org/sqlite`), task state, local media paths, range-capable HTTP streaming, and all yt-dlp process execution. The client never executes commands or opens local paths.

## Providers

`VideoProvider` supplies `Name`, `SyncFavorites`, `ListVideos`, and `Download`. `MockProvider` serves stable demo records and generates playable test MP4/M4A with ffmpeg while reporting copied-byte progress. `YtDlpProvider` is selected with `POCKETVIDEO_PROVIDER=ytdlp`; Go runs yt-dlp for video/audio downloads, parses progress, supports yt-dlp's native continuation, writes into the configured download directory, and applies the configured proxy. Favorite listing requires Bilibili login and currently returns an actionable error. A future `BBDownProvider` can implement the same interface without changing routes or UI.

## Persistence and media

SQLite migrations are ordered and transactional; each successful step appends its version to `schema_version`. The service stores favorites, videos, tasks, settings, and media status. Mock media is copied to the configured download directory with persisted byte offsets, and `/api/task/{id}/resume` continues from that offset. Local media is served through `/api/media/{id}` with `http.ServeFile`, which supports byte ranges; audio uses `?kind=audio`. Cache directory is reserved for artwork and metadata.

## Automatic synchronization

The service starts a goroutine that reads `auto_sync` and `sync_minutes` from settings. When enabled, it waits for the configured interval and refreshes favorites and videos through the active provider. Failures are logged and retried without crashing the service; disabling automatic sync stops provider calls. Manual `POST /api/sync` uses the same synchronization function.

## PocketJS Runtime / .pj migration

The frontend is a browser-only Vite application and contains no Node or Electron integration. A PocketJS adapter can replace `fetch` with the runtime's HTTP and media APIs while retaining JSON contracts and responsive React views. Package the static frontend and Go service in the future `.pj` format, then map the service's data root and media capabilities to PJOS permissions. Keep provider and SQL layers behind the service boundary. Validate `<video>` playback and filesystem permissions against the eventual runtime; browsers currently stream only through the local HTTP API.
