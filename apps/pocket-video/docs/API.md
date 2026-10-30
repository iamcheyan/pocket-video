# HTTP API

All endpoints use JSON except media. Errors use `{"error":{"code":"...","message":"..."}}`.

- `GET /api/health` → `{"ok":true,"provider":"mock","mock":true,"database":"sqlite","schema_version":1}`
- `GET /api/favorites` → array of `{id,title,video_count,last_sync_at}`
- `GET /api/videos?q=offline&limit=20&offset=0` → `{items:[video],limit,offset}`. Video fields include `id,favorite_id,title,creator,duration,has_video,has_audio`.
- `GET /api/videos/{id}` → one video; 404 when absent.
- `GET /api/task` → array of `{id,video_id,kind,status,downloaded_bytes,total_bytes,progress}`.
- `POST /api/download` body `{"video_id":"BV1DEMO001","kind":"video"}` → 202 `{id,status}`. Invalid kinds/JSON return 400.
- `POST /api/task/{id}/pause` → task pauses at a progress checkpoint. `POST /api/task/{id}/resume` → resumes a paused task.
- `POST /api/sync` → `{favorites,videos}` or 503 if the provider cannot sync.
- `GET /api/settings` → settings object. `PUT /api/settings` accepts the same object; language must be `en`, `zh-CN`, or `ja`; theme must be `system`, `light`, or `dark`.
- `GET /api/media/{id}` → video bytes with HTTP Range support; 404 if no downloaded video is recorded.

Start with `make backend` (port 8080). The Vite development server proxies `/api` to this service.
