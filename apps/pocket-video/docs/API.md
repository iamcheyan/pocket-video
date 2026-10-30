# HTTP API

The Go service exposes one JSON API. The frontend reaches it through `api.ts`
only; the Vite dev server proxies `/api` to `127.0.0.1:8080`. Media and artwork
are the only non-JSON responses.

## Conventions

- Request and response bodies are UTF-8 JSON.
- Errors always use the same envelope and an HTTP status:

  ```json
  { "error": { "code": "invalid_state", "message": "task is not running" } }
  ```

- `id` values are Bilibili BV ids (`BV1DEMO00x` in mock mode).
- Lists are ordered deterministically (`videos` by id, `favorites` by title,
  `task` newest first) so screenshots and tests are stable.
- `422` is not used; validation failures are `400`, missing resources `404`,
  state conflicts `409`, provider failures `503`.

## Endpoints

### `GET /api/health`

```json
{"ok":true,"provider":"mock","mock":true,"mode":"demo","authenticated":true,"database":"sqlite","schema_version":1}
```

### `GET /api/account`

Describes the active adapter without exposing credentials:

```json
{"mode":"demo","provider":"mock","authenticated":true,"display_name":"Demo Account","collections":4}
```

`mode` is `demo` for `MockProvider` and `live` for `YtDlpProvider`. With the live
provider, `authenticated` is `false` until a login flow exists.

### `GET /api/favorites`

```json
[{"id":"fav-pjs","title":"PocketJS Deep Dives","video_count":4,"last_sync_at":"2026-10-30T09:16:47Z"}]
```

`video_count` is recomputed from `videos` after every sync, so it always matches
what the library can list.

### `GET /api/videos`

Query parameters:

| Name | Meaning |
| --- | --- |
| `q` | Case-insensitive match on title or creator |
| `favorite_id` | Restrict to one favorite folder |
| `limit` | 1–100, default 50 |
| `offset` | Zero-based, default 0 |

```json
{
  "items": [{
    "id": "BV1DEMO002",
    "favorite_id": "fav-pjs",
    "title": "Offline First: A Practical Guide",
    "creator": "PocketVideo Lab",
    "duration": 428,
    "cover": "/api/cover/BV1DEMO002",
    "has_video": true,
    "has_audio": false
  }],
  "limit": 50,
  "offset": 0,
  "total": 12,
  "downloaded": 1
}
```

- `total` counts all matches; `downloaded` counts the whole library's videos with
  a finished video file (used by the home statistics).
- `cover` is a URL or an empty string. Local filesystem paths are never returned.
- `has_video` / `has_audio` are the only media-state flags a client needs.

### `GET /api/videos/{id}`

One video object in the same shape, or `404 not_found`.

### `GET /api/task` and `GET /api/task/{id}`

```json
[{
  "id": 2,
  "video_id": "BV1DEMO002",
  "title": "Offline First: A Practical Guide",
  "kind": "video",
  "status": "done",
  "downloaded_bytes": 151508,
  "total_bytes": 151508,
  "progress": 100,
  "updated_at": "2026-10-30 09:27:37"
}]
```

`title` is joined from `videos` so a queue row renders without a second lookup.
`progress` is always derived from the byte counters (`0` when the total is
unknown). `error` appears only for failed tasks.

### `POST /api/download`

```json
{"video_id":"BV1DEMO002","kind":"video"}
```

- `kind` must be `video` or `audio`; both are required.
- First request: `202 {"id":1,"status":"queued"}`.
- While an active task exists for the same video/kind: `200` with
  `{"id":1,"status":"running","existing":true}`. A paused task is resumed instead
  of duplicated.
- Unknown video: `404 not_found`; malformed body: `400 invalid_request`.

### `POST /api/task/{id}/pause`

Cancels the worker and waits until the persisted state is `paused`, then returns
`{"status":"paused"}`. `409 invalid_state` when the task is not running.

### `POST /api/task/{id}/resume`

Restarts a `queued` or `paused` task from its stored byte offset and returns
`{"status":"running"}`. Other states return `409 invalid_state`.

### `DELETE /api/task/{id}`

Removes a queue entry. Completed media stays in the library; partial files of
paused/failed tasks are deleted. Returns `204`. A running task must be paused
first (`409 invalid_state`).

### `POST /api/sync`

Refreshes favorites and videos through the provider and recomputes counters:

```json
{"favorites":4,"videos":12,"synced_at":"2026-10-30T09:30:00Z"}
```

Provider failures return `503 provider_unavailable` with the reason (for example
the missing Bilibili login). Automatic sync uses the same path driven by
`auto_sync` + `sync_minutes`.

### `GET /api/settings` and `PUT /api/settings`

```json
{
  "language": "en",
  "theme": "system",
  "cache_dir": "/srv/pocketvideo/cache",
  "download_dir": "/srv/pocketvideo/downloads",
  "sync_minutes": 60,
  "auto_sync": false,
  "proxy": ""
}
```

Validation on `PUT`:

- `language` ∈ `en`, `zh-CN`, `ja`
- `theme` ∈ `system`, `light`, `dark`
- `sync_minutes` ∈ 1–10080
- `cache_dir` and `download_dir` must be absolute paths

Invalid payloads return `400 invalid_settings` and leave the stored row untouched.
The response is the stored settings object.

### `GET /api/media/{id}`

Streams the downloaded video with `Accept-Ranges: bytes` (media seeking and
resume rely on this). `?kind=audio` streams the audio track. Missing media or an
unknown id returns `404 media_not_found`; an invalid kind returns `400`.

### `GET /api/cover/{id}`

Serves the JPEG extracted from a downloaded mock video, with
`Cache-Control: public, max-age=86400`. `404 cover_not_found` when no artwork has
been generated yet.

## Task state machine

| From | Event | To |
| --- | --- | --- |
| — | `POST /api/download` | `queued` → `running` |
| `running` | worker finishes | `done` |
| `running` | provider error | `failed` (with `error`) |
| `running` | `POST …/pause` | `paused` |
| `paused` | `POST …/resume` | `running` |
| `paused` | `DELETE` | removed + partial file cleaned |
| `running` | service restart | `paused` (resume hint) |

## Not part of the contract

- Authenticated Bilibili sync, cookie import, and playlist discovery.
- Deleting library videos (removing a completed task keeps the media).
- Multi-user accounts or remote access. The service binds to localhost by
  default and is not designed to be exposed to a network.
