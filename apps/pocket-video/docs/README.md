# PocketVideo documentation

Start with the [project README](../README.md), then pick a topic:

| Document | Contents |
| --- | --- |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Layer boundaries, request flow, provider adapters, media pipeline, PocketJS/PJOS portability |
| [API.md](API.md) | REST contract, payloads, validation, error envelope, task state machine |
| [DATABASE.md](DATABASE.md) | SQLite schema, migrations, seed data, status transitions |
| [DEVELOPMENT.md](DEVELOPMENT.md) | Local workflow, mock catalog, adding a provider, i18n/layout conventions |
| [PACKAGING.md](PACKAGING.md) | Planned `.pj` package layout and PJOS runtime mapping |
| [ROADMAP.md](ROADMAP.md) | What exists, what is next, known gaps |
| [SCREENSHOTS.md](SCREENSHOTS.md) | Verified capture gallery and reproduction command |

## Commands

```sh
make backend        # Go service on :8080 (mock provider)
make frontend       # Vite dev server on :5173
make test           # Go tests, frontend tests, production build, screenshot verification
make screenshots    # start missing servers, seed demo downloads, recapture the gallery
make verify         # verify images and layout without recapturing
```

Environment variables:

| Variable | Meaning | Default |
| --- | --- | --- |
| `POCKETVIDEO_DATA` | SQLite, cache, and download root | `./data` in `backend/` |
| `POCKETVIDEO_ADDR` | Listen address | `:8080` |
| `POCKETVIDEO_PROVIDER` | `mock` or `ytdlp` | `mock` |
| `POCKETVIDEO_PROXY` | Proxy for the live provider at startup | empty |
| `YT_DLP_BINARY` | yt-dlp executable | `yt-dlp` from `PATH` |
| `CHROME_BIN` | Browser for screenshots | auto-detected |
