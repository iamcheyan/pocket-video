package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err = migrate(db); err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	root := t.TempDir()
	s := &Server{db: db, provider: MockProvider{}, root: root, cancels: map[int64]context.CancelFunc{}, providerConfig: ProviderDownloadConfig{OutputDir: filepath.Join(root, "downloads")}, workerStates: map[int64]string{}}
	s.seed()
	t.Cleanup(func() { db.Close() })
	return s
}
func TestMigrationAndMockProvider(t *testing.T) {
	s := testServer(t)
	var n int
	if e := s.db.QueryRow("SELECT count(*) FROM videos").Scan(&n); e != nil || n != len(demoVideos) {
		t.Fatalf("seeded videos=%d err=%v", n, e)
	}
	vs, e := (MockProvider{}).ListVideos(context.Background(), Query{Q: "offline"})
	if e != nil || len(vs) != 1 {
		t.Fatalf("mock search=%d err=%v", len(vs), e)
	}
	var favorites int
	if e := s.db.QueryRow("SELECT count(*) FROM favorites").Scan(&favorites); e != nil || favorites != len(demoFavorites) {
		t.Fatalf("seeded favorites=%d err=%v", favorites, e)
	}
}
func TestAPIRoutesAndSettings(t *testing.T) {
	s := testServer(t)
	m := http.NewServeMux()
	s.routes(m)
	for _, path := range []string{"/api/health", "/api/favorites", "/api/videos", "/api/videos/BV1DEMO001", "/api/task", "/api/settings"} {
		r := httptest.NewRecorder()
		m.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if r.Code != 200 {
			t.Errorf("%s status %d: %s", path, r.Code, r.Body.String())
		}
	}
	body := `{"language":"ja","theme":"dark","cache_dir":"/cache","download_dir":"/media","sync_minutes":15,"auto_sync":true,"proxy":""}`
	r := httptest.NewRecorder()
	m.ServeHTTP(r, httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body)))
	if r.Code != 200 {
		t.Fatalf("settings write: %d %s", r.Code, r.Body.String())
	}
	var lang, theme string
	if e := s.db.QueryRow("SELECT language,theme FROM settings").Scan(&lang, &theme); e != nil || lang != "ja" || theme != "dark" {
		t.Fatalf("settings persistence %q %q %v", lang, theme, e)
	}
}
func request(m http.Handler, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRecorder()
	m.ServeHTTP(r, httptest.NewRequest(method, path, strings.NewReader(body)))
	return r
}
func waitTask(t *testing.T, s *Server, id int64, want string) Task {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var x Task
		e := s.db.QueryRow("SELECT id,video_id,kind,status,downloaded_bytes,total_bytes,COALESCE(error,'') FROM download_tasks WHERE id=?", id).Scan(&x.ID, &x.VideoID, &x.Kind, &x.Status, &x.DownloadedBytes, &x.TotalBytes, &x.Error)
		if e != nil {
			t.Fatal(e)
		}
		if x.Status == want {
			return x
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("task %d did not reach %s", id, want)
	return Task{}
}
func TestDownloadPauseResumeTransitions(t *testing.T) {
	s := testServer(t)
	m := http.NewServeMux()
	s.routes(m)
	r := request(m, "POST", "/api/download", `{"video_id":"BV1DEMO001","kind":"video"}`)
	if r.Code != 202 {
		t.Fatalf("download status %d %s", r.Code, r.Body.String())
	}
	var id int64
	if e := s.db.QueryRow("SELECT id FROM download_tasks").Scan(&id); e != nil {
		t.Fatal(e)
	}
	var before int64
	var status, taskError string
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		_ = s.db.QueryRow("SELECT downloaded_bytes,status,COALESCE(error,'') FROM download_tasks WHERE id=?", id).Scan(&before, &status, &taskError)
		if before > 0 || status == "failed" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if before == 0 {
		t.Fatalf("download made no progress: status=%s error=%s", status, taskError)
	}
	r = request(m, "POST", "/api/task/"+strconv.FormatInt(id, 10)+"/pause", "")
	if r.Code != 200 {
		t.Fatalf("pause status %d: %s", r.Code, r.Body.String())
	}
	paused := waitTask(t, s, id, "paused")
	if paused.DownloadedBytes == 0 || paused.DownloadedBytes >= paused.TotalBytes {
		t.Fatalf("pause did not retain a partial byte offset: %+v", paused)
	}
	r = request(m, "POST", "/api/task/"+strconv.FormatInt(id, 10)+"/resume", "")
	if r.Code != 200 {
		t.Fatalf("resume status %d: %s", r.Code, r.Body.String())
	}
	done := waitTask(t, s, id, "done")
	if done.DownloadedBytes != done.TotalBytes || done.DownloadedBytes <= paused.DownloadedBytes {
		t.Fatalf("resume did not continue to completion: paused=%d done=%d/%d", paused.DownloadedBytes, done.DownloadedBytes, done.TotalBytes)
	}
	var path string
	if e := s.db.QueryRow("SELECT video_path FROM videos WHERE id='BV1DEMO001'").Scan(&path); e != nil {
		t.Fatal(e)
	}
	if st, e := os.Stat(path); e != nil || st.Size() != done.TotalBytes {
		t.Fatalf("completed file stat=%v err=%v", st, e)
	}
	req := httptest.NewRequest("GET", "/api/media/BV1DEMO001", nil)
	req.Header.Set("Range", "bytes=0-11")
	media := httptest.NewRecorder()
	m.ServeHTTP(media, req)
	if media.Code != http.StatusPartialContent || media.Header().Get("Content-Range") == "" || len(media.Body.Bytes()) != 12 {
		t.Fatalf("video range status=%d range=%q bytes=%d", media.Code, media.Header().Get("Content-Range"), media.Body.Len())
	}
	audio := request(m, "POST", "/api/download", `{"video_id":"BV1DEMO003","kind":"audio"}`)
	if audio.Code != http.StatusAccepted {
		t.Fatalf("audio task status=%d %s", audio.Code, audio.Body.String())
	}
	var audioID int64
	if err := s.db.QueryRow("SELECT id FROM download_tasks WHERE kind='audio'").Scan(&audioID); err != nil {
		t.Fatal(err)
	}
	audioDone := waitTask(t, s, audioID, "done")
	if audioDone.TotalBytes == 0 || audioDone.DownloadedBytes != audioDone.TotalBytes {
		t.Fatalf("audio task incomplete: %+v", audioDone)
	}
	var audioPath string
	if err := s.db.QueryRow("SELECT audio_path FROM videos WHERE id='BV1DEMO003'").Scan(&audioPath); err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(audioPath); err != nil || st.Size() != audioDone.TotalBytes {
		t.Fatalf("audio file stat=%v err=%v", st, err)
	}
}
func TestProviderFailureDoesNotFakeSuccess(t *testing.T) {
	s := testServer(t)
	s.provider = YtDlpProvider{Binary: "/definitely/missing/yt-dlp", OutputDir: s.providerConfig.OutputDir}
	m := http.NewServeMux()
	s.routes(m)
	r := request(m, "POST", "/api/download", `{"video_id":"BV1DEMO002","kind":"video"}`)
	if r.Code != 202 {
		t.Fatalf("queue status %d", r.Code)
	}
	var id int64
	_ = s.db.QueryRow("SELECT id FROM download_tasks").Scan(&id)
	failed := waitTask(t, s, id, "failed")
	if failed.Error == "" {
		t.Fatal("failure reason missing")
	}
	var path sql.NullString
	_ = s.db.QueryRow("SELECT video_path FROM videos WHERE id='BV1DEMO002'").Scan(&path)
	if path.Valid {
		t.Fatalf("failed task recorded fake media %q", path.String)
	}
}
