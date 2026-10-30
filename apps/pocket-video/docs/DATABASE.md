# Database

SQLite via `modernc.org/sqlite` (pure Go, no cgo). The database file is
`${POCKETVIDEO_DATA:-./data}/pocketvideo.db`; startup enables WAL, foreign keys,
and applies ordered migrations.

## Tables

```sql
favorites(
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  video_count INTEGER NOT NULL DEFAULT 0,
  last_sync_at TEXT
)

videos(
  id TEXT PRIMARY KEY,
  favorite_id TEXT NOT NULL REFERENCES favorites(id),
  title TEXT NOT NULL,
  creator TEXT NOT NULL,
  duration INTEGER NOT NULL,      -- seconds, from provider metadata
  cover TEXT,                     -- cached JPEG path, never sent to clients
  video_path TEXT,                -- downloaded file path, never sent to clients
  audio_path TEXT                 -- downloaded file path, never sent to clients
)
CREATE INDEX videos_title_idx ON videos(title)

download_tasks(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  video_id TEXT NOT NULL,
  kind TEXT NOT NULL CHECK(kind IN ('video','audio')),
  status TEXT NOT NULL CHECK(status IN ('queued','running','paused','done','failed')),
  downloaded_bytes INTEGER NOT NULL DEFAULT 0,
  total_bytes INTEGER NOT NULL DEFAULT 0,
  error TEXT,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
)

settings(
  id INTEGER PRIMARY KEY CHECK(id=1),
  language TEXT NOT NULL,
  theme TEXT NOT NULL,
  cache_dir TEXT NOT NULL,
  download_dir TEXT NOT NULL,
  sync_minutes INTEGER NOT NULL,
  auto_sync INTEGER NOT NULL,     -- 0/1
  proxy TEXT NOT NULL
)

schema_version(version INTEGER NOT NULL)
```

Notes:

- `cover`, `video_path`, and `audio_path` hold local paths for the service only.
  API responses expose `/api/cover/{id}` and `/api/media/{id}` URLs plus
  `has_video`/`has_audio` booleans instead.
- `settings` is a singleton row. `PUT /api/settings` performs an upsert after
  validation, so a rejected payload cannot partially write.
- `download_tasks.video_id` intentionally has no foreign key: a task must remain
  readable even if a future sync removes a video from the catalog.

## Migrations

`migrate.go` keeps an ordered list of functions. Each runs inside its own
transaction together with the `schema_version` insert, so a failed step rolls
back completely and the next startup retries from the last good version.
Rerunning migrations is idempotent; `automation_test.go` asserts that the version
and row count stay at 1 after a second run.

The current schema is version 1. Adding a column means appending a function, not
editing migration 1, so existing installations upgrade in place.

## Seed data

On first start with the mock provider:

- four favorite folders (`fav-pjs`, `fav-audio`, `fav-ui`, `fav-music`)
- twelve videos (`BV1DEMO001`–`BV1DEMO012`) with mixed English, Chinese, and
  Japanese titles
- one settings row using `<data>/cache` and `<data>/downloads`

A live (`ytdlp`) provider is never seeded with demo rows. `video_count` is
recomputed from `videos` after seeding and after every sync, so the favorites
page can never drift.

## Task state persistence

- Progress is written after every 64 KiB chunk (`downloaded_bytes`, `total_bytes`,
  `updated_at`). That is the resume granularity.
- Pause sets `paused` only after the worker stops, so the stored offset matches
  the bytes on disk.
- A failed provider records the error message for the queue and detail views.
- On startup, rows left in `running` become `paused` with
  `error='service restarted; resume to continue'`; the media file and byte offset
  are preserved for resume.
- Deleting a completed task removes only the queue row. Deleting a paused/failed
  task also removes its partial media and any `.source` file used by the mock
  provider.

## Files on disk

```text
<data>/
  pocketvideo.db            SQLite database (WAL sidecars while running)
  cache/<VIDEO_ID>.jpg      extracted cover art
  downloads/<VIDEO_ID>.video.mp4
  downloads/<VIDEO_ID>.audio.m4a
```

The download directory is configurable in settings and used by workers; the cache
directory stores artwork. Demo media is generated on demand and is not committed
to the repository.

## Backup and reset

- Backup: stop the service, copy the whole data directory (database plus
  downloads) — WAL checkpoints on clean shutdown.
- Reset the demo: stop the service and delete the data directory. The next start
  recreates schema, seed data, and settings.
