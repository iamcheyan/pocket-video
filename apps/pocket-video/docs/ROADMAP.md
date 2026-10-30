# Roadmap

## Implemented

- Browser React/TypeScript frontend and Go JSON API.
- SQLite schema creation, seeded demo collection, settings persistence, search, task progress/pause/resume, media endpoint and mock provider.
- English, Simplified Chinese, and Japanese labels; system/light/dark themes; responsive navigation, keyboard focus indicators, touch-sized controls.
- yt-dlp provider entry point and user-visible unavailable errors.

## Not implemented

- Bilibili login and authenticated favorite extraction. No account/cookies were available; yt-dlp sync returns an explicit error.
- Real media retrieval through yt-dlp is currently not connected to task execution. The current development task writes a small labeled placeholder file, not playable video/audio. It is not valid to claim real download playback works.
- Automatic periodic sync scheduler, proxy application to yt-dlp, and honoring per-user cache/download directory settings.
- Gamepad navigation. Standard browser keyboard focus and Tab/Enter navigation are available; dedicated arrow-key/gamepad mapping remains future work.
- PocketJS Runtime and `.pj` packaging; documented migration path only.
- Cross-restart resume of an active task. Persisted task state and downloaded byte count exist; the in-process demo worker is not recovered after service restart.

## Known runtime limits

Browsers can play only media formats they support; future PocketJS filesystem and codec permissions need runtime validation. The demo media placeholder is not a playable stream. See `SCREENSHOTS.md` for tested screenshot viewports and the commands used.
