# PocketVideo

## Quick start

Requirements: Go, Node.js, npm, and `ffmpeg` (used to generate playable deterministic demo media).

Start one terminal per service:

```sh
cd apps/pocket-video/backend && go mod download && go run .
cd apps/pocket-video/frontend && npm install && npm run dev
```

Open `http://localhost:5173`. The backend defaults to Mock mode and stores SQLite data under `./data`. For the real download provider, set `POCKETVIDEO_PROVIDER=ytdlp` when starting the backend; install yt-dlp separately. The provider runs yt-dlp only from Go, downloads media into the configured download directory, and applies the saved proxy. Authenticated Bilibili favorite sync requires login and currently reports a clear unavailable error. Do not place cookies or credentials in the repository.

Run `make test` from the repository root for Go unit/integration coverage, i18n key parity, and TypeScript/Vite production build. Run `make backend` and `make frontend` in separate terminals as launch helpers.

## Layout

`apps/pocket-video/frontend` is the browser client. `backend` is the Go service, SQLite owner and sole command execution boundary. `docs` contains architecture/API/database/roadmap and screenshot index. `tests` describes the test entry points.

See [SCREENSHOTS.md](SCREENSHOTS.md) for the Chromium screenshot file list, pixel dimensions, and reproduction command.
