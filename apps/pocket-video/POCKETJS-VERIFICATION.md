# POCKETJS VERIFICATION — pocket-video

Round: PocketJS refactor. Device acceptance is a later pass.

## Shape

- Pocket app: `apps/pocket-video/pocket/`
- Frozen React reference: `apps/pocket-video/frontend/` (design reference; not extended).
- TSX files in the port: **6** (`main.tsx`, `app.tsx`, `ui.tsx`, three screens).
- Native tree only: `View`/`Text` from `@pocketjs/framework/solid/components`; no DOM, iframe, texture screenshot, or `dangerouslySetInnerHTML`.

## Manifest

- id: `dev.pocket-video.pocket`
- entry: `src/main.tsx`
- output: `pocket-video`
- viewport: `360×240 @2x` → physical `720×480`
- capabilities: `text.glyphs.baked`, `input.buttons`
- target: `rocknix-h700-kms`

## Admission / compile

```sh
cd apps/pocket-video/pocket
./setup-sdk.sh
SDK=${POCKETJS_SDK:-$HOME/development/.pocketjs-star-drift-sdk}
bun "$SDK/tools/pocket.ts" check \
  --runtime-config "$POCKET_VERIFY/pjs/profiles/pocket.runtime-config.ts" \
  --target rocknix-h700-kms --manifest pocket.json
bun "$SDK/tools/pocket.ts" compile \
  --runtime-config "$POCKET_VERIFY/pjs/profiles/pocket.runtime-config.ts" \
  --target rocknix-h700-kms --manifest pocket.json --output out.pocket
```

Admission: `sha256:dfc6707fe12e9ddfe5b6d27835a5eff59acf15abd31dc5da0bff6383570fc8b8`; TypeScript passes for the app module graph. Package: `out.pocket`, sha256 `7548161fe8ce3148414b8d3d24932bf6c7eb26fe9a48fc726373f6dd4d727849`.

## Frames

Rendered with the SDK wasm core at `renderScaled(2)` (720×480), not a browser:

```sh
cd "$POCKET_VERIFY"
bun tools/render-pocket.ts \
  --dir repos/pocket-video/apps/pocket-video/pocket \
  --script repos/pocket-video/apps/pocket-video/pocket/frames.json \
  --out repos/pocket-video/docs/pocket
```

| file | state | sha256 (first 16) |
|---|---|---|
| `docs/pocket/screen-home.png` | home | `c110dc06d40e1b84` |
| `docs/pocket/screen-home-focus.png` | home-focus | `ebc3647d7bc089ab` |
| `docs/pocket/screen-detail.png` | detail | `7832d854e77e33b9` |
| `docs/pocket/screen-detail-action.png` | detail-action | `2e558af54222bf60` |
| `docs/pocket/screen-settings.png` | settings | `30813f0669e8d361` |
| `docs/pocket/screen-settings-toggled.png` | settings-toggled | `29751281500d3e0c` |

All six files have distinct bytes; the renderer rejects duplicate frames. The
tape is: move focus, open the item, run the first detail action, switch to
settings, move a row, toggle it.

Screens: Offline archive/queue, player detail, archive settings. Interactions: PREV/PLAY/NEXT/QUEUE; progress and queue badge change, no media playback.

## Adapter

- **fixture** — this port runs from JSON baked into the package and labels the
  header `fixture`. `client.ts` also contains the `http` and `bridge` adapters
  behind the same `list()/get()/action()` surface.
- `http` uses `@pocketjs/framework/net-api`; valid only in the host dev profile.
- `bridge` uses `host.__pvc(JSON.stringify(request))`; the runtime forwards to
  the Go daemon on `127.0.0.1:<port>` with a hard timeout.
- Device gap: videod bridge plus yt-dlp/mpv. The device image has neither.
- Daemon contract: /api/videos, /api/task, /api/favorites; providers mock.

## Not done / open questions

- Device run, saves/annotations and the bridge itself are out of this round.
- Real media/audio, OAuth, emulator, engine and system writes stay unimplemented
  where the original product already marked them mock or unavailable.
- Long lists use the six-item fixture surface; paging/virtual lists are a
  follow-up once the bridge provides paginated data.
