# Roadmap

## Implemented

- Browser React/TypeScript frontend and Go JSON API; `make backend` serves the current project build.
- SQLite schema creation, seeded demo collection, settings persistence, search, task progress/pause/resume, and byte-range media streaming.
- Mock downloads produce playable MP4 video and M4A audio via `ffmpeg`; tasks persist byte offsets and resume appending from the last saved byte.
- English, Simplified Chinese, and Japanese; system/light/dark themes; responsive layouts and keyboard, arrow-key, and Gamepad API navigation.
- yt-dlp download execution is routed through Go; missing binary, login, network, or unavailable media returns a failed task with its error recorded.

## Not implemented

- Authenticated Bilibili favorite extraction and login flow. No account/cookies were available; yt-dlp sync returns an explicit error.
- The mock provider uses generated test media rather than copyrighted Bilibili media.
- Automatic periodic sync scheduler. Proxy and download directory settings are applied to yt-dlp and downloaded media; cache directory remains reserved for metadata/artwork.
- PocketJS Runtime and `.pj` packaging; migration path is documented only.

## Known runtime limits

Browsers can play only media formats they support; future PocketJS filesystem and codec permissions need runtime validation. The demo media placeholder is not a playable stream. See `SCREENSHOTS.md` for tested screenshot viewports and the commands used.
