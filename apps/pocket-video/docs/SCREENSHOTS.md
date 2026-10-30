# Screenshots

Captured from the running Vite application at `http://127.0.0.1:5173`, which proxies `/api` to the Go service on `127.0.0.1:8080`. Every image is a real Chromium capture of a live page. All 19 files have distinct SHA-256 hashes and verified PNG dimensions; `make test` fails if a duplicate appears.

## Route and state captures (1280×720 CSS viewport, 1× device scale)

| File | Output pixels | Live state |
|---|---:|---|
| `screenshots/home-1280x720.png` | 1280×720 | Home, English, light |
| `screenshots/favorites-1280x720.png` | 1280×720 | Favorites, English, light |
| `screenshots/videos-1280x720.png` | 1280×720 | Video list, English, light |
| `screenshots/detail-1280x720.png` | 1280×720 | Video detail, English, light |
| `screenshots/queue-1280x720.png` | 1280×720 | Download queue, English, light |
| `screenshots/settings-1280x720.png` | 1280×720 | Settings, English, light |
| `screenshots/language-zh-1280x720.png` | 1280×720 | Settings, Simplified Chinese |
| `screenshots/language-ja-1280x720.png` | 1280×720 | Settings, Japanese |
| `screenshots/theme-dark-1280x720.png` | 1280×720 | Home, English, dark |

## Responsive viewport captures

| File | Output pixels | Live state |
|---|---:|---|
| `screenshots/view-320-240.png` | 320×240 | Video list, English, light |
| `screenshots/view-480-320.png` | 480×320 | Video list, English, light |
| `screenshots/view-640-480.png` | 640×480 | Video list, English, light |
| `screenshots/view-720-720.png` | 720×720 | Video list, English, light |
| `screenshots/view-1280-720.png` | 1280×720 | Favorites, Japanese, light |
| `screenshots/view-1920-1080.png` | 1920×1080 | Home, English, light |

## DPI captures (same dark English detail state at four scales)

| File | Output pixels | Device scale |
|---|---:|---|
| `screenshots/dpi-1x.png` | 1280×720 | 1× |
| `screenshots/dpi-1_5x.png` | 1920×1080 | 1.5× |
| `screenshots/dpi-2x.png` | 2560×1440 | 2× |
| `screenshots/dpi-3x.png` | 3840×2160 | 3× |

## Reproduction

Start the backend and frontend, then run one command:

```sh
make screenshots
```

The script `scripts/capture-screenshots.py` opens each URL below in headless Chromium, writes to `docs/screenshots/`, verifies each PNG header against its expected pixel size, and asserts that 19 files have 19 unique SHA-256 hashes:

```text
http://127.0.0.1:5173/?theme=light&lang=en#home
http://127.0.0.1:5173/?theme=light&lang=en#favorites
http://127.0.0.1:5173/?theme=light&lang=en#videos
http://127.0.0.1:5173/?theme=light&lang=en#detail
http://127.0.0.1:5173/?theme=light&lang=en#queue
http://127.0.0.1:5173/?theme=light&lang=en#settings
http://127.0.0.1:5173/?theme=light&lang=zh-CN#settings
http://127.0.0.1:5173/?theme=light&lang=ja#settings
http://127.0.0.1:5173/?theme=dark&lang=en#home
http://127.0.0.1:5173/?theme=light&lang=ja#favorites   (1280×720)
```

Viewport and DPI captures use Chromium `--window-size` plus `--force-device-scale-factor`; see `CAPTURES` in the script for every exact command argument.

The 320×240 workflow was additionally exercised interactively in Chromium: drawer → Videos → detail → download → queue → Settings, with no horizontal overflow.
