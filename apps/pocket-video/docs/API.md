# HTTP API

All endpoints use JSON except media. Errors use `{"error":{"code":"...","message":"..."}}`.

- `GET /api/health` → `{"ok":true,"provider":"mock","mock":true,"database":"sqlite","schema_version":1}`
- `GET /api/favorites` → array of `{id,title,video_count,last_sync_at}`
- `GET /api/videos?q=offline&limit=20&offset=0` → `{items:[video],limit,offset}`. Video fields include `id,favorite_id,title,creator,duration,has_video,has_audio`.
- `GET /api/videos/{id}` → one video; 404 when absent.
- `GET /api/task` → array of `{id,video_id,kind,status,downloaded_bytes,total_bytes,progress}`.
- `POST /api/download` body `{"video_id":"BV1DEMO001","kind":"video"}` → 202 `{id,status}`. Invalid kinds/JSON return 400.
- `POST /api/task/{id}/pause` → waits for a stable paused state and returns `{status:"paused"}` after retaining downloaded bytes. `POST /api/task/{id}/resume` → resumes the existing partial media from the persisted offset.
- `POST /api/sync` → refreshes favorites and videos and returns current `{favorites,videos}` counts, or 503 if the provider cannot sync. Saved `auto_sync` and `sync_minutes` settings also drive repeated synchronization in the service.
- `GET /api/settings` → settings object. `PUT /api/settings` accepts the same object; language must be `en`, `zh-CN`, or `ja`; theme must be `system`, `light`, or `dark`. Download directory and proxy affect provider execution; cache directory is reserved for cached artwork/metadata.
- `GET /api/media/{id}` → video bytes with HTTP Range support. Add `?kind=audio` for audio media; 404 if the requested downloaded media is unavailable.
Start with `make backend` (port 8080). The Vite development server proxies `/api` to this service.
