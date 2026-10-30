# PocketVideo

## Quick start

Requirements: Go 1.24+, Node.js 20+, npm. Install dependencies and run two terminals:

```sh
cd apps/pocket-video/backend && go mod download && go run .
cd apps/pocket-video/frontend && npm install && npm run dev
```

Open `http://localhost:5173`. The service defaults to deterministic Mock mode and creates SQLite data under `./data`. To select the real-provider error path, run the backend with `POCKETVIDEO_PROVIDER=ytdlp`; yt-dlp must be installed for real process execution, but authenticated Bilibili sync is not implemented. Never add cookies to this repository.

Run `make test` at repository root for backend tests, i18n key parity, and frontend TypeScript/Vite build. Run `make backend` and `make frontend` as convenient wrappers.

## Layout

`apps/pocket-video/frontend` is the browser client. `backend` is the Go service, SQLite owner and sole command execution boundary. `docs` contains architecture/API/database/roadmap and screenshot index. `tests` describes the test entry points.

See [SCREENSHOTS.md](SCREENSHOTS.md) for the Chromium screenshot file list, pixel dimensions, and reproduction command.
