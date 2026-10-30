# Roadmap

## Implemented

- Browser React/TypeScript frontend and Go JSON API; `make backend` serves the current project build.
- Transactional, ordered SQLite migrations with a recorded schema version, seeded demo collection, settings persistence, search, task progress/pause/resume, and byte-range media streaming.
- Mock downloads produce playable MP4 video and M4A audio via `ffmpeg`; tasks persist byte offsets and resume appending from the last saved byte.
- English, Simplified Chinese, and Japanese; system/light/dark themes; responsive layouts and keyboard, arrow-key, and Gamepad API navigation.
- yt-dlp download execution is routed through Go; missing binary, login, network, or unavailable media returns a failed task with its error recorded.
- Automatic favorite synchronization in the Go service. When enabled, it checks the stored `sync_minutes` interval and refreshes favorites and videos; disabling it stops provider calls.

## Not implemented

- Authenticated Bilibili favorite extraction and login flow. No account/cookies were available; yt-dlp sync returns an explicit error.
- The mock provider uses generated test media rather than copyrighted Bilibili media. Cache and download directories persist as settings; download media uses the configured download directory.
- PocketJS Runtime and `.pj` packaging; migration path is documented only.

## Known runtime limits

Browsers can play only media formats they support; future PocketJS filesystem and codec permissions need runtime validation. Mock media is generated H.264/AAC video or AAC audio and is playable in current Chromium. See `SCREENSHOTS.md` for tested screenshot viewports and the commands used.
