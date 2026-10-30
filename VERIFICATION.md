# Verification — pocket-video @ 2026-10-31

- Verdict: PASS
- Commit: `e17ae8cc12fc25da564604a34cc8f3b840405910` (favicon fix + screenshots; this document is the following commit)
- Remote: `e17ae8cc12fc25da564604a34cc8f3b840405910`
- Runtime: `python3 scripts/capture-screenshots.py --serve` → Go service `127.0.0.1:8080` (mock provider, SQLite) + Vite `127.0.0.1:5173`; production build also verified with `npm run build`
- Toolchain: go 1.26.0, node v22.23.3, Chrome 155.0.8059.84, ffmpeg present

## Test gate
| command | result | notes |
|---|---|---|
| `go test ./...` | PASS | 12 Go tests (API, automation, provider, worker), 0 failures |
| `node --test --experimental-strip-types tests/*.test.ts` | PASS | 24 tests, 0 failures |
| `npm run build` | PASS | `tsc -b && vite build` |
| `python3 scripts/capture-screenshots.py --verify` | PASS | 21 PNGs, 21 unique SHA-256, PNG sizes match, no horizontal overflow at 320×240 / 480×320 / 640×480 / 720×720 / 1280×720 |

`make test` runs all four gates; green in ~12 s (with the stack already warm).

## Screenshot matrix
Captured this run from the live app (Vite + Go service, temporary data directory, seeded
video + audio downloads). `apps/pocket-video/docs/screenshots/manifest.json` records the
capture timestamp (`2026-10-31T01:32:24+0900`), capture URL, Chrome path, and per-file
SHA-256.

| screen | viewport | theme | file | note |
|---|---|---|---|---|
| collection | 1280×720 | see name | `apps/pocket-video/docs/screenshots/collection-1280x720.png` | sha256 88179076f8c788ca |
| detail | 1280×720 | see name | `apps/pocket-video/docs/screenshots/detail-1280x720.png` | sha256 d44bc4fcf6b41ecf |
| detail-audio | 1280×720 | see name | `apps/pocket-video/docs/screenshots/detail-audio-1280x720.png` | sha256 e8133991c622c497 |
| dpi | 1920×1080 | see name | `apps/pocket-video/docs/screenshots/dpi-1_5x.png` | sha256 76c09491764ba6ba |
| dpi | 1280×720 | see name | `apps/pocket-video/docs/screenshots/dpi-1x.png` | sha256 6998387585b0ae4c |
| dpi | 2560×1440 | see name | `apps/pocket-video/docs/screenshots/dpi-2x.png` | sha256 3a6c49ca3de67023 |
| dpi | 3840×2160 | see name | `apps/pocket-video/docs/screenshots/dpi-3x.png` | sha256 04bb94e654f1d523 |
| favorites | 1280×720 | see name | `apps/pocket-video/docs/screenshots/favorites-1280x720.png` | sha256 651b787afcb82d7b |
| home | 1280×720 | see name | `apps/pocket-video/docs/screenshots/home-1280x720.png` | sha256 96784bab92e78e11 |
| language-ja | 1280×720 | see name | `apps/pocket-video/docs/screenshots/language-ja-1280x720.png` | sha256 57def29dc7b7d103 |
| language-zh | 1280×720 | see name | `apps/pocket-video/docs/screenshots/language-zh-1280x720.png` | sha256 f66cdb53022b72b1 |
| queue | 1280×720 | see name | `apps/pocket-video/docs/screenshots/queue-1280x720.png` | sha256 ea243c00788029a5 |
| settings | 1280×720 | see name | `apps/pocket-video/docs/screenshots/settings-1280x720.png` | sha256 548be2a6cd8deb6c |
| theme-dark | 1280×720 | see name | `apps/pocket-video/docs/screenshots/theme-dark-1280x720.png` | sha256 58e91a815edbb6d9 |
| videos | 1280×720 | see name | `apps/pocket-video/docs/screenshots/videos-1280x720.png` | sha256 62948cd9516210cc |
| view-1280 | 1280×720 | see name | `apps/pocket-video/docs/screenshots/view-1280-720.png` | sha256 5ff1e18dd2590736 |
| view-1920 | 1920×1080 | see name | `apps/pocket-video/docs/screenshots/view-1920-1080.png` | sha256 ca1a4aa4b7d28463 |
| view-320 | 320×240 | see name | `apps/pocket-video/docs/screenshots/view-320-240.png` | sha256 c89f1c44b5312ade |
| view-480 | 480×320 | see name | `apps/pocket-video/docs/screenshots/view-480-320.png` | sha256 3614af818d74f40f |
| view-640 | 640×480 | see name | `apps/pocket-video/docs/screenshots/view-640-480.png` | sha256 cdaf47fe5ff0b97a |
| view-720 | 720×720 | see name | `apps/pocket-video/docs/screenshots/view-720-720.png` | sha256 e1eadbe0c6161800 |

Count: 21 screenshots, 21 unique hashes (0 duplicates). `tools/check_shots.py`: all PNGs valid,
dimensions match the declared CSS viewport × DPI, minimum per-image RGB stddev well above the
blank threshold, no blank frames.

Runtime probe (`tools/probe.mjs`) against the live stack for `/home`, `/favorites`,
`/videos`, `/favorites/fav-pjs`, `/detail/BV1DEMO002`, `/queue`, `/settings` at 1280×720:
HTTP 200, no console/page errors, no failed or 4xx/5xx requests, no horizontal overflow.
`/home` also clean at 320×240.

## Fixed
- Every page loaded with a console error: Chrome requested `/favicon.ico`, the app had no icon
  and Vite answered 404 (`Failed to load resource: ... 404`). → added
  `apps/pocket-video/frontend/public/favicon.svg` and `<link rel="icon" type="image/svg+xml">`
  in `index.html` → probe is now clean at 1280×720 and 320×240; the icon is also emitted into
  `dist/` by the production build.
- Added `docs/screenshots/manifest.json` (written by the capture script on every run) so a
  committed screenshot set carries its own capture timestamp and SHA-256 provenance instead of
  relying on commit dates.

## Blocked / open questions
- None. The live yt-dlp/Bilibili provider was not exercised (mock mode is the documented demo
  path and needs no account); network degradation paths belong to that provider.

## Not verified here
- 掌机真机（RG34XX / PJOS-LINUX）未验证；本轮只在本机 headless Chrome 里跑真实进程。
- Real Bilibili downloads / yt-dlp network paths; DRM or copy-protected media.
