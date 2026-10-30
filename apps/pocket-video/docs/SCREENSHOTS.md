# Screenshots

Captured from the running Vite application at `http://127.0.0.1:5173` (proxying the Go service on `127.0.0.1:8081`). All images are real Chromium captures of the live app. Verify dimensions with `file apps/pocket-video/docs/screenshots/*.png`.

## Route and state captures (1280×720 CSS viewport, 1× device scale)

| File | Output pixels | Description |
|---|---:|---|
| `screenshots/home-1280x720.png` | 1280×720 | English home page |
| `screenshots/favorites-1280x720.png` | 1280×720 | Favorites page |
| `screenshots/videos-1280x720.png` | 1280×720 | Video list page |
| `screenshots/detail-1280x720.png` | 1280×720 | Video detail page |
| `screenshots/queue-1280x720.png` | 1280×720 | Download queue page |
| `screenshots/settings-1280x720.png` | 1280×720 | Settings page (English) |
| `screenshots/language-zh-1280x720.png` | 1280×720 | Settings page in Simplified Chinese |
| `screenshots/language-ja-1280x720.png` | 1280×720 | Settings page in Japanese |
| `screenshots/theme-dark-1280x720.png` | 1280×720 | Dark theme applied and persisted |

## Responsive viewport captures (1× device scale)

| File | Output pixels | Description |
|---|---:|---|
| `screenshots/view-320-240.png` | 320×240 | Compact home viewport |
| `screenshots/view-480-320.png` | 480×320 | Small landscape viewport |
| `screenshots/view-640-480.png` | 640×480 | VGA viewport |
| `screenshots/view-720-720.png` | 720×720 | Square viewport |
| `screenshots/view-1280-720.png` | 1280×720 | Desktop viewport |
| `screenshots/view-1920-1080.png` | 1920×1080 | Full HD viewport |

## DPI captures (1280×720 CSS viewport)

| File | Output pixels | Description |
|---|---:|---|
| `screenshots/dpi-1x.png` | 1280×720 | 1× device scale |
| `screenshots/dpi-1_5x.png` | 1920×1080 | 1.5× device scale |
| `screenshots/dpi-2x.png` | 2560×1440 | 2× device scale |
| `screenshots/dpi-3x.png` | 3840×2160 | 3× device scale |

## Reproduction

```sh
cd apps/pocket-video/backend && POCKETVIDEO_ADDR=127.0.0.1:8081 go run .
cd apps/pocket-video/frontend && npm run dev
chromium --headless=new --disable-gpu --no-sandbox --hide-scrollbars --force-device-scale-factor=1 --window-size=320,240 --screenshot=apps/pocket-video/docs/screenshots/view-320-240.png http://127.0.0.1:5173/
```

The 320×240 flow was verified in Chromium: open the drawer, select Videos, open a detail, start a download, wait for DONE, reopen the drawer, and reach Settings with no horizontal overflow.