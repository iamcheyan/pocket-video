# Database

The Go service uses SQLite via `modernc.org/sqlite`, so builds do not require cgo. The database is `${POCKETVIDEO_DATA:-./data}/pocketvideo.db`; startup enables WAL and creates schema version 1 idempotently.

- `schema_version(version)` tracks the active schema. Add future upgrades as ordered Go migration functions before updating the recorded version.
- `favorites(id,title,video_count,last_sync_at)` stores synchronized collections.
- `videos(id,favorite_id,title,creator,duration,cover,video_path,audio_path)` stores metadata and cached media references; title has an index.
- `download_tasks(id,video_id,kind,status,downloaded_bytes,total_bytes,error,updated_at)` constrains kinds to video/audio and states to queued/running/paused/done/failed.
- `settings(id,language,theme,cache_dir,download_dir,sync_minutes,auto_sync,proxy)` stores the user's preferences.

The demo downloader writes progress and supports pause/resume, retaining the byte count. Resume continues from the saved count in the mock task simulation. No user cookies or credentials are stored.
