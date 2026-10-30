# Database

The Go service uses SQLite via `modernc.org/sqlite`, so builds do not require cgo. The database is `${POCKETVIDEO_DATA:-./data}/pocketvideo.db`; startup enables WAL and creates schema version 1 idempotently.

- `schema_version(version)` tracks active schema version 1. Startup creates each table idempotently; future versions should use ordered migrations and update this value only after successful migration.
- `favorites(id,title,video_count,last_sync_at)` stores synchronized collections.
- `videos(id,favorite_id,title,creator,duration,cover,video_path,audio_path)` stores metadata and cached media references; title has an index.
- `download_tasks(id,video_id,kind,status,downloaded_bytes,total_bytes,error,updated_at)` constrains kinds to video/audio and states to queued/running/paused/done/failed.
- `settings(id,language,theme,cache_dir,download_dir,sync_minutes,auto_sync,proxy)` stores the user's preferences.

Mock tasks generate playable MP4/M4A media using `ffmpeg`, copy it into the configured download directory while persisting byte counts, and resume from the last durable byte offset. At startup, abandoned `running` tasks are reset to `queued`; the user can explicitly resume a queued task. No user cookies or credentials are stored.
