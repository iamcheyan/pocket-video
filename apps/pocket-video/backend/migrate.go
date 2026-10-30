package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

var migrations = []func(*sql.Tx) error{
	func(tx *sql.Tx) error {
		_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS favorites(id TEXT PRIMARY KEY,title TEXT NOT NULL,video_count INTEGER NOT NULL DEFAULT 0,last_sync_at TEXT);
CREATE TABLE IF NOT EXISTS videos(id TEXT PRIMARY KEY,favorite_id TEXT NOT NULL,title TEXT NOT NULL,creator TEXT NOT NULL,duration INTEGER NOT NULL,cover TEXT,video_path TEXT,audio_path TEXT,FOREIGN KEY(favorite_id) REFERENCES favorites(id));
CREATE INDEX IF NOT EXISTS videos_title_idx ON videos(title);
CREATE TABLE IF NOT EXISTS download_tasks(id INTEGER PRIMARY KEY AUTOINCREMENT,video_id TEXT NOT NULL,kind TEXT NOT NULL CHECK(kind IN ('video','audio')),status TEXT NOT NULL CHECK(status IN ('queued','running','paused','done','failed')),downloaded_bytes INTEGER NOT NULL DEFAULT 0,total_bytes INTEGER NOT NULL DEFAULT 0,error TEXT,updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE IF NOT EXISTS settings(id INTEGER PRIMARY KEY CHECK(id=1),language TEXT NOT NULL,theme TEXT NOT NULL,cache_dir TEXT NOT NULL,download_dir TEXT NOT NULL,sync_minutes INTEGER NOT NULL,auto_sync INTEGER NOT NULL,proxy TEXT NOT NULL);`)
		return err
	},
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON; CREATE TABLE IF NOT EXISTS schema_version(version INTEGER NOT NULL);`); err != nil {
		return err
	}
	var version int
	if err := db.QueryRow(`SELECT COALESCE(MAX(version),0) FROM schema_version`).Scan(&version); err != nil {
		return err
	}
	for next := version; next < len(migrations); next++ {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if err = migrations[next](tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %d: %w", next+1, err)
		}
		if _, err = tx.Exec(`INSERT INTO schema_version(version) VALUES(?)`, next+1); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
		log.Printf("applied schema migration %d", next+1)
	}
	return nil
}

func (s *Server) startAutoSync(ctx context.Context) {
	go func() {
		for {
			poll := s.autoSyncEvery
			if poll <= 0 {
				poll = 15 * time.Second
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(poll):
			}
			var enabled int
			if err := s.db.QueryRow(`SELECT auto_sync FROM settings WHERE id=1`).Scan(&enabled); err != nil || enabled != 1 {
				continue
			}
			if err := s.syncFavorites(ctx); err != nil {
				log.Printf("automatic sync failed: %v", err)
			}
		}
	}()
}

func (s *Server) syncFavorites(ctx context.Context) error {
	fs, err := s.provider.SyncFavorites(ctx)
	if err != nil {
		return err
	}
	for _, f := range fs {
		if _, err = s.db.Exec(`INSERT INTO favorites(id,title,video_count,last_sync_at) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET title=excluded.title,video_count=excluded.video_count,last_sync_at=excluded.last_sync_at`, f.ID, f.Title, f.VideoCount, f.LastSyncAt); err != nil {
			return err
		}
	}
	vs, err := s.provider.ListVideos(ctx, Query{})
	if err != nil {
		return err
	}
	for _, v := range vs {
		if _, err = s.db.Exec(`INSERT INTO videos(id,favorite_id,title,creator,duration,cover) VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET title=excluded.title,creator=excluded.creator,duration=excluded.duration,cover=excluded.cover`, v.ID, v.FavoriteID, v.Title, v.Creator, v.Duration, v.Cover); err != nil {
			return err
		}
	}
	s.refreshFavoriteCounts()
	return nil
}

// refreshFavoriteCounts makes the stored counters match the video table, so the
// favorites page is always consistent with what the library can show.
func (s *Server) refreshFavoriteCounts() {
	_, _ = s.db.Exec(`UPDATE favorites SET video_count=(SELECT count(*) FROM videos WHERE videos.favorite_id=favorites.id)`)
}
