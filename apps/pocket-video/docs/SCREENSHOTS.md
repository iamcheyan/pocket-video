# Screenshots

All captures come from the live Vite app at `http://127.0.0.1:5173`, proxying
`/api` to the Go service on `127.0.0.1:8080`. The capture script starts missing
servers with a temporary data directory, seeds one video and one audio download,
and captures real states — including the completed queue, the offline player
poster, and the custom audio transport. 21 files have distinct SHA-256 hashes and
verified PNG dimensions; `make test` fails on a missing, duplicated, wrong-sized,
or horizontally overflowing state.

## Provenance

`manifest.json` in this directory is rewritten on every capture with the timestamp, capture
URL, Chrome binary and SHA-256 of each PNG, so a run can be tied to its artefacts.

## Screens and states (1280×720 CSS viewport, 1× scale)

| File | Output | State |
| --- | ---: | --- |
| `screenshots/home-1280x720.png` | 1280×720 | Home, English, light |
| `screenshots/favorites-1280x720.png` | 1280×720 | Four favorite folders, English, light |
| `screenshots/collection-1280x720.png` | 1280×720 | One folder filtered, English, light |
| `screenshots/videos-1280x720.png` | 1280×720 | Library with a downloaded video and cover |
| `screenshots/detail-1280x720.png` | 1280×720 | Video detail, offline poster + play |
| `screenshots/detail-audio-1280x720.png` | 1280×720 | Audio detail with custom transport |
| `screenshots/queue-1280x720.png` | 1280×720 | Two completed downloads |
| `screenshots/settings-1280x720.png` | 1280×720 | Settings, English, light |
| `screenshots/language-zh-1280x720.png` | 1280×720 | Settings, 简体中文 |
| `screenshots/language-ja-1280x720.png` | 1280×720 | Settings, 日本語 |
| `screenshots/theme-dark-1280x720.png` | 1280×720 | Home, English, dark |

## Responsive viewports

| File | Output | Viewport | State |
| --- | ---: | --- | --- |
| `screenshots/view-320-240.png` | 320×240 | 掌机 / 最小支持尺寸 | Compact list, English |
| `screenshots/view-480-320.png` | 480×320 | 小屏掌机 | Compact list, English |
| `screenshots/view-640-480.png` | 640×480 | 阅读器 / 小屏 | Compact list, English |
| `screenshots/view-720-720.png` | 720×720 | 方形设备 | Compact list, English |
| `screenshots/view-1280-720.png` | 1280×720 | 桌面 | Favorites, 日本語 |
| `screenshots/view-1920-1080.png` | 1920×1080 | 全高清桌面 | Home, English |

## DPI scaling

The same dark English detail state, rasterized at four device scales:

| File | Output pixels | Device scale |
| --- | ---: | ---: |
| `screenshots/dpi-1x.png` | 1280×720 | 1× |
| `screenshots/dpi-1_5x.png` | 1920×1080 | 1.5× |
| `screenshots/dpi-2x.png` | 2560×1440 | 2× |
| `screenshots/dpi-3x.png` | 3840×2160 | 3× |

## Reproduction

```sh
make screenshots        # recapture everything, starting servers if needed
make verify             # verify images and layouts without recapturing
```

`scripts/capture-screenshots.py`:

1. resolves Chrome (`CHROME_BIN`, `chromium`, `google-chrome`, or a Playwright
   cache), and starts the Go service / Vite dev server only if they are down;
2. seeds one video and one audio download so queue and player states are real;
3. captures each URL with an isolated `--user-data-dir` so runs never hang on a
   locked profile;
4. verifies PNG signatures, expected pixel dimensions, and unique hashes;
5. loads the video list at 320×240, 480×320, 640×480, 720×720, and 1280×720 and
   asserts the app reports no horizontal overflow
   (`<html data-overflow="false">`).

Example URLs (theme and language are URL parameters so a capture is reproducible):

```text
http://127.0.0.1:5173/?theme=light&lang=en#home
http://127.0.0.1:5173/?theme=light&lang=en#favorites
http://127.0.0.1:5173/?theme=light&lang=en#favorites/fav-pjs
http://127.0.0.1:5173/?theme=light&lang=en#videos
http://127.0.0.1:5173/?theme=light&lang=en#detail/BV1DEMO002
http://127.0.0.1:5173/?theme=light&lang=en#detail/BV1DEMO004
http://127.0.0.1:5173/?theme=light&lang=en#queue
http://127.0.0.1:5173/?theme=light&lang=en#settings
http://127.0.0.1:5173/?theme=light&lang=zh-CN#settings
http://127.0.0.1:5173/?theme=light&lang=ja#settings
http://127.0.0.1:5173/?theme=dark&lang=en#home
http://127.0.0.1:5173/?theme=dark&lang=en#detail/BV1DEMO002   (DPI set, 1×–3×)
```

The 320×240 flow was also exercised as an interaction path: menu → Videos →
detail → download → queue → Settings, with the focus ring and Esc back behaviour
verified by keyboard.
