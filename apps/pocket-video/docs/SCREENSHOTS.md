# Screenshots

Captured from the running Vite application at `http://127.0.0.1:5173` with Chromium headless. Viewport files use `--window-size=WIDTH,HEIGHT --force-device-scale-factor=1`; DPI files use a 1280×720 CSS viewport and the indicated `--force-device-scale-factor`. Each viewport capture is a real application screenshot. Verify dimensions with `file apps/pocket-video/docs/screenshots/*.png`.

| File | Output pixels | Description |
|---|---:|---|
| `screenshots/view-320-240.png` | 320×240 | Compact home viewport |
| `screenshots/view-480-320.png` | 480×320 | Small landscape viewport |
| `screenshots/view-640-480.png` | 640×480 | VGA viewport |
| `screenshots/view-720-720.png` | 720×720 | Square viewport |
| `screenshots/view-1280-720.png` | 1280×720 | Desktop viewport |
| `screenshots/view-1920-1080.png` | 1920×1080 | Full HD viewport |
| `screenshots/dpi-1_5x.png` | 1920×1080 | 1280×720 viewport at 1.5× device scale |
| `screenshots/dpi-2x.png` | 2560×1440 | 1280×720 viewport at 2× device scale |
| `screenshots/dpi-3x.png` | 3840×2160 | 1280×720 viewport at 3× device scale |

The app routes (home, favorites, video list, video detail, queue, settings), Chinese and Japanese localization, and dark theme were exercised in managed Chromium against the live Vite/Go services. The screenshots above prioritize accurate pixel dimensions and DPI. To reproduce the route-specific visual captures, navigate through the UI after starting both services, then run Chromium with `--screenshot=FILE` and the route's viewport size.

Example command:

```sh
chromium --headless=new --disable-gpu --no-sandbox --hide-scrollbars --force-device-scale-factor=1 --window-size=320,240 --screenshot=apps/pocket-video/docs/screenshots/view-320-240.png http://127.0.0.1:5173/
```
