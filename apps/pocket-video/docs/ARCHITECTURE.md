# Architecture

PocketVideo keeps the browser as a thin TypeScript/React client. It calls a Go HTTP service using JSON. The service is the only business and persistence boundary: it owns SQLite (`modernc.org/sqlite`), task state, local media paths, range-capable HTTP streaming, and all yt-dlp process execution. The client never executes commands or opens local paths.

## Providers

`VideoProvider` supplies `Name`, `SyncFavorites`, `ListVideos`, and `Download`. `MockProvider` serves stable demo records and deterministic progress for offline development. `YtDlpProvider` is selected with `POCKETVIDEO_PROVIDER=ytdlp`; its operations return a useful missing-login message if favorites cannot be fetched. It receives IDs from the service and launches yt-dlp only from Go. Bilibili authenticated favorite extraction is not yet implemented. A future `BBDownProvider` can implement the same interface without changing routes or UI.

## Persistence and media

SQLite migrations run at startup and store favorites, videos, tasks, settings, and the schema version. The development downloader writes a clearly labeled placeholder media file; it is useful to demonstrate task flow but is not playable video content. Real media is served through `/api/media/{id}` with `http.ServeFile`, which supports byte ranges. Settings directory fields are currently preferences; the demo writes into the configured service data root.

## PocketJS Runtime / .pj migration

The frontend is a browser-only Vite application and contains no Node or Electron integration. A PocketJS adapter can replace `fetch` with the runtime's HTTP and media APIs while retaining JSON contracts and responsive React views. Package the static frontend and Go service in the future `.pj` format, then map the service's data root and media capabilities to PJOS permissions. Keep provider and SQL layers behind the service boundary. Validate `<video>` playback and filesystem permissions against the eventual runtime; browsers currently stream only through the local HTTP API.
