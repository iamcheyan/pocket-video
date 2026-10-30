# Development

## Prerequisites

| Tool | Why |
| --- | --- |
| Go 1.24+ | Backend service; module already pins the SQLite driver |
| Node.js 22+ / npm | Frontend build, dev server, unit tests |
| ffmpeg | Demo media generation and cover extraction |
| Chrome/Chromium | Screenshot capture and layout verification |
| yt-dlp | Only for `POCKETVIDEO_PROVIDER=ytdlp` |

## Running

```sh
make backend            # :8080, mock provider, data in backend/data
make frontend           # :5173, proxies /api to :8080
```

Reset the demo by stopping the backend and deleting `apps/pocket-video/backend/data`.

Useful variants:

```sh
POCKETVIDEO_DATA=/tmp/pv POCKETVIDEO_ADDR=127.0.0.1:8080 make backend
POCKETVIDEO_PROVIDER=ytdlp YT_DLP_BINARY=/usr/bin/yt-dlp make backend
```

## Where to change what

| Task | Files |
| --- | --- |
| Add or change an endpoint | `backend/api.go`, route table in `backend/main.go`, `docs/API.md` |
| Add demo content | `backend/provider.go` (`demoFavorites`, `demoVideos`) |
| Change download behavior | `backend/worker.go`, `backend/download.go` |
| Add a schema field | new migration in `backend/migrate.go`, docs, tests |
| Add a screen | `frontend/src/pages/`, route in `router.ts`, `App.tsx` switch |
| Add UI copy | `frontend/src/i18n.ts` (all three locales; types enforce parity) |
| Add an API call | `frontend/src/api.ts` only — never `fetch` inside a page |
| Change layout/theme | `frontend/src/style.css`, `theme.ts` |

## The mock catalog

`demoFavorites` and `demoVideos` in `provider.go` are the single source of demo
content. They are deterministic so tests, screenshots, and documentation stay
stable. Adding a video means adding one struct entry; `SyncFavorites` recomputes
`video_count`, so no other bookkeeping is needed.

Demo media is generated on demand by `generateDemoMedia` (a muted gradient clip
plus a tone) and artwork by `generateDemoPoster` (frame at 1s). Missing ffmpeg
fails the download with a clear error; cover generation is best-effort.

## Adding a provider

1. Implement the interface in `types.go`:

   ```go
   type VideoProvider interface {
       Name() string
       Account(context.Context) (Account, error)
       SyncFavorites(context.Context) ([]Favorite, error)
       ListVideos(context.Context, Query) ([]Video, error)
       Download(context.Context, DownloadRequest, string, int64, func(Progress)) (Result, error)
   }
   ```

2. Respect the `Progress` contract: report monotonically increasing bytes and the
   total size when known; the worker persists every callback.
3. Honour `ctx.Done()` so pause works, and start from the provided `offset` when
   the format allows resume. Set `Result.Path` only when a file exists.
4. Select it in `main.go` (environment variable or settings) and add a test with
   a fake provider, following `automation_test.go`.
5. Update `Account` so the UI can show mode, provider, and authentication state.

Never execute commands from the frontend; provider work belongs in Go.

## Tests

```sh
make test                                               # everything
cd apps/pocket-video/backend && go test ./...            # API, DB, providers
cd apps/pocket-video && npm --prefix frontend run build  # typecheck + bundle
cd apps/pocket-video && node --test --experimental-strip-types tests/*.test.ts
python3 scripts/capture-screenshots.py --verify          # gallery + layouts
```

Guidelines:

- API tests use `httptest` and a real in-memory SQLite database; do not mock the
  database.
- Frontend tests target pure modules (api client, router, format, theme, i18n)
  and run without a DOM. UI behavior is covered by the screenshot pipeline and
  manual device checks.
- A test must never claim success for work it did not do: provider failures are
  asserted to leave `video_path` empty (`TestProviderFailureDoesNotFakeSuccess`).

## i18n rules

- English is the source dictionary; the other locales are typed against it, so
  `tsc` fails when a key is missing.
- Use `{count}` / `{time}` placeholders with `formatCount`; do not concatenate
  sentences in components.
- `tests/i18n.test.ts` checks parity, non-empty values, placeholder survival,
  and the two-line hero copy.

## Layout and input rules

- No native select/dialog styling assumptions: controls are custom (theme
  segmented buttons, toggles, custom audio transport) so they render identically
  on handhelds.
- Touch targets grow under `max-height: 460px`; the compact list mode replaces
  cards with thumbnail rows.
- Do not hardcode viewport assumptions; test 320×240 first when adding UI.
- Keep the focus ring intact: every interactive element must be a real
  `button`/`a`/`input`, labeled with `aria-label` when icon-only.
- `data-overflow="true"` on `<html>` means the page needs horizontal scrolling;
  the screenshot pipeline treats that as a failure.

## Debugging

- Backend logs go to stdout; a failed task stores its provider error in
  `download_tasks.error` and exposes it through the API.
- `curl -s localhost:8080/api/health` is the fastest liveness check.
- `GET /api/video/{id}` shows whether media is recorded; `/api/media/{id}` with
  `Range: bytes=0-99` should return `206`.
- If screenshots hang, delete a stale Chrome profile: the script already uses a
  temporary `--user-data-dir` per capture and a per-capture timeout.
- Use `POCKETVIDEO_DATA=$(mktemp -d)` to isolate an experiment from your demo
  database.
