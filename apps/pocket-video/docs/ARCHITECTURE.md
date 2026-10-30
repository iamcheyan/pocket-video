# Architecture

PocketVideo is a three-layer reference application. Each layer has one owner and
one direction of dependency:

```text
┌──────────────────────────────────────────────────────────┐
│ PocketJS surface (React + TypeScript, Vite build)        │
│  pages/ ── components/ ── i18n ── theme ── router         │
│                 │                                         │
│             api.ts  (only module allowed to call HTTP)    │
└─────────────────┼────────────────────────────────────────┘
                  │ JSON over HTTP, media by URL
┌─────────────────▼────────────────────────────────────────┐
│ Go service                                               │
│  api.go ── worker.go ── migrate.go ── provider.go        │
│     │         │            │             │               │
│   SQLite   task state   schema      VideoProvider         │
│                                          │               │
│                                  MockProvider            │
│                                  YtDlpProvider → yt-dlp  │
└──────────────────────────────────────────────────────────┘
```

The UI never executes commands, opens SQLite, or receives filesystem paths.
Everything the client knows about the device arrives through the REST contract.

## Backend layers

| File | Responsibility |
| --- | --- |
| `main.go` | Server struct, startup, route table, CORS, JSON helpers, seed |
| `types.go` | API-facing structs and the `VideoProvider` interface |
| `api.go` | HTTP handlers and validation; no business logic beyond translation |
| `worker.go` | Download workers, pause/resume state, media/cover persistence |
| `provider.go` | `MockProvider` catalog and `YtDlpProvider` metadata |
| `download.go` | Provider download implementations, progress parsing, demo media |
| `migrate.go` | Ordered migrations, sync, auto-sync loop |

`Server` owns the database, the active adapter, one cancellation function per
running task, and the last observed worker state. All writes go through this
struct, so there is a single place to reason about concurrency.

## Request flow

1. The client calls `api.ts`, which prefixes the configured base (`/api` by
   default, or `globalThis.__POCKETVIDEO_API__` for embedded runtimes).
2. Handlers validate input, query SQLite, and return JSON. Errors use one
   envelope: `{"error":{"code","message"}}`.
3. The client stores views in React state; it never derives persistence. The
   service is the source of truth for settings, library content, and tasks.
4. A one-second poll of `/api/task` drives progress. When a task completes, the
   client refreshes the library once.

## Provider adapters

```go
type VideoProvider interface {
    Name() string
    Account(context.Context) (Account, error)
    SyncFavorites(context.Context) ([]Favorite, error)
    ListVideos(context.Context, Query) ([]Video, error)
    Download(context.Context, DownloadRequest, string, int64, func(Progress)) (Result, error)
}
```

- **MockProvider** is deterministic: four favorite folders, twelve videos, no
  network. Media is generated with ffmpeg from a muted gradient clip and a tone,
  so downloads, pause/resume, playback, and artwork are real behaviors rather
  than simulated counters.
- **YtDlpProvider** runs yt-dlp from Go: `--continue`, proxy support, progress
  parsing, video/audio selection, output discovery. Metadata sync needs a login
  that does not exist yet, and returns an explicit error instead of fake data.
- **BBDownProvider** (future) only needs to implement this interface; routes,
  SQLite, workers, and the UI remain unchanged.

## Media and artwork

Downloaded files live in the configured download directory. SQLite stores the
paths, but `json:"-"` keeps them out of API responses. Clients use:

- `GET /api/media/{id}` — byte-range streaming for `<video>`/`<audio>`
- `GET /api/media/{id}?kind=audio` — audio track
- `GET /api/cover/{id}` — cached artwork extracted with ffmpeg after a successful
  mock video download

Video paths are never returned; `has_video`/`has_audio` and the two URL families
are the whole contract. This is what lets PJOS replace HTTP streaming with a
runtime media handle without changing page components.

## Task lifecycle

```text
queued ──► running ──► done
   │          │  │
   │          │  └──► failed (provider error, recorded in `error`)
   │          └──► paused ──► resume ──► running
   └──► (service restart: running becomes paused)
```

- Exactly one active task per `(video_id, kind)`; a duplicate request resumes a
  paused task or returns the existing one.
- Pause waits for the worker to observe cancellation, so the persisted byte
  offset is stable before the response returns.
- Removing a task is only allowed when it is not running; partial files are
  deleted, completed media stays in the library.
- On startup, abandoned `running` rows become `paused` with a resume hint.

## Frontend structure

| Area | Files |
| --- | --- |
| API boundary | `api.ts` (typed client, `ApiError`, URL helpers) |
| Routing | `router.ts` (hash routes, deep links, round-trip helpers) |
| Screens | `pages/HomePage`, `FavoritesPage`, `VideosPage`, `DetailPage`, `QueuePage`, `SettingsPage` |
| Shell | `App.tsx`, `components/Sidebar`, `components/TopBar` |
| Media | `components/VideoGrid`, `components/AudioPlayer` |
| Platform | `theme.ts`, `i18n.ts`, `navigation.ts`, `format.ts`, `types.ts` |

### i18n

`i18n.ts` defines English first and types the other dictionaries against it, so a
missing key fails `tsc`. Runtime tests additionally assert key parity, non-empty
copy, placeholders, and language negotiation. No UI string is hardcoded in a
component; content (video titles, creators) is data.

### Theme

`theme.ts` resolves `system | light | dark` to a concrete theme, writes
`data-theme` and `color-scheme`, and subscribes to
`prefers-color-scheme` changes while in system mode. `main.tsx` applies the theme
before React mounts to avoid a flash on handhelds.

### Responsive and DPI strategy

Layout is fluid: `clamp()` typography, grid columns that collapse at 850px and
560px, and a **compact list mode** under `max-height: 460px` that turns cards into
thumbnail rows with larger tap targets. Under `max-height: 320px` the decorative
hero is removed so content wins. DPI is handled by vector layout and rem sizing;
1×–3× captures verify that text and controls remain legible when rasterized at
higher densities.

### Input

`navigation.ts` provides a single spatial focus ring for keyboard arrows and the
Gamepad API (D-pad, left stick, A confirm, B back). Enter activates the focused
control, Esc goes back, `/` and ⌘/Ctrl+K focus search. All icon buttons carry
`aria-label`s and the active nav item uses `aria-current="page"`.

## Portability to PocketJS / PJOS

- The frontend is a static Vite build with no Electron or Node integration.
- All business access goes through `api.ts`; a runtime adapter can replace
  `fetch` and the media URLs with PocketJS IPC without touching pages.
- The Go service is a single binary with one data root and no cgo dependency.
- `GET /api/account` already abstracts "who is the provider and is it
  authenticated", which a PJOS permission prompt can populate.
- See [PACKAGING.md](PACKAGING.md) for the planned `.pj` layout.

## Deliberate trade-offs

- **Polling instead of websockets/SSE.** One-second task polling is adequate for
  local downloads, trivial to port, and keeps the runtime contract small.
- **Synthetic demo media.** The mock provider must be deterministic and
  redistributable, so it generates its own clips instead of shipping Bilibili
  content. Demo durations can differ from the catalog metadata.
- **HTTP media streaming.** The browser needs URLs; PJOS can later expose a
  native media handle behind the same two endpoints.
- **No login.** Rather than storing cookies without a supported flow, the live
  provider reports an actionable error and the mock provider carries the demo.
