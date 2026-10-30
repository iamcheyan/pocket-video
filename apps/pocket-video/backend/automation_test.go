package main

import (
	"context"
	"database/sql"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestMigrationsAreIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "schema.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = migrate(db); err != nil {
		t.Fatal(err)
	}
	if err = migrate(db); err != nil {
		t.Fatalf("second migration run: %v", err)
	}
	var version, count int
	if err = db.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow("SELECT count(*) FROM schema_version").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if version != 1 || count != 1 {
		t.Fatalf("schema version=%d rows=%d", version, count)
	}
	for _, table := range []string{"favorites", "videos", "download_tasks", "settings"} {
		var n int
		if err = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&n); err != nil || n != 1 {
			t.Fatalf("table %s missing", table)
		}
	}
}

type countingProvider struct{ calls int }

func (p *countingProvider) Name() string { return "counting" }
func (p *countingProvider) Account(context.Context) (Account, error) {
	return Account{Mode: "demo", Provider: "counting", Authenticated: true, DisplayName: "Counting"}, nil
}
func (p *countingProvider) SyncFavorites(context.Context) ([]Favorite, error) {
	p.calls++
	return []Favorite{{ID: "auto", Title: "Automatic favorites", VideoCount: 1, LastSyncAt: time.Now().UTC().Format(time.RFC3339)}}, nil
}
func (p *countingProvider) ListVideos(context.Context, Query) ([]Video, error) {
	return []Video{{ID: "auto-video", FavoriteID: "auto", Title: "Scheduled refresh", Creator: "PocketVideo Lab", Duration: 60}}, nil
}
func (p *countingProvider) Download(context.Context, DownloadRequest, string, int64, func(Progress)) (Result, error) {
	return Result{}, nil
}

func TestAutomaticSyncRefreshesFavorites(t *testing.T) {
	s := testServer(t)
	provider := &countingProvider{}
	s.provider = provider
	s.autoSyncEvery = 10 * time.Millisecond
	if _, err := s.db.Exec("UPDATE settings SET auto_sync=1,sync_minutes=1 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.startAutoSync(ctx)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var n int
		_ = s.db.QueryRow("SELECT count(*) FROM favorites WHERE id='auto'").Scan(&n)
		if n == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("automatic synchronization did not refresh favorites")
}
func TestDisabledAutomaticSyncDoesNothing(t *testing.T) {
	s := testServer(t)
	provider := &countingProvider{}
	s.provider = provider
	s.autoSyncEvery = 10 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.startAutoSync(ctx)
	time.Sleep(60 * time.Millisecond)
	if provider.calls != 0 {
		t.Fatalf("provider called %d times while automatic sync was disabled", provider.calls)
	}
}
func init() { http.DefaultClient.Timeout = 5 * time.Second }
