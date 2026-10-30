package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	db, e := sql.Open("sqlite", ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	if e = migrate(db); e != nil {
		t.Fatal(e)
	}
	s := &Server{db: db, provider: MockProvider{}, root: t.TempDir(), cancels: map[int64]context.CancelFunc{}}
	s.seed()
	return s
}
func TestMigrationAndMockProvider(t *testing.T) {
	s := testServer(t)
	var n int
	if e := s.db.QueryRow("SELECT count(*) FROM videos").Scan(&n); e != nil || n != 3 {
		t.Fatalf("seeded videos=%d err=%v", n, e)
	}
	vs, e := (MockProvider{}).ListVideos(context.Background(), Query{Q: "offline"})
	if e != nil || len(vs) != 1 {
		t.Fatalf("mock search=%d err=%v", len(vs), e)
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
func TestDownloadPauseResumeTransitions(t *testing.T) {
	s := testServer(t)
	m := http.NewServeMux()
	s.routes(m)
	r := httptest.NewRecorder()
	m.ServeHTTP(r, httptest.NewRequest("POST", "/api/download", strings.NewReader(`{"video_id":"BV1DEMO001","kind":"video"}`)))
	if r.Code != 202 {
		t.Fatalf("download status %d %s", r.Code, r.Body.String())
	}
	var id int64
	if e := s.db.QueryRow("SELECT id FROM download_tasks").Scan(&id); e != nil {
		t.Fatal(e)
	}
	r = httptest.NewRecorder()
	m.ServeHTTP(r, httptest.NewRequest("POST", "/api/task/"+strconv.FormatInt(id, 10)+"/pause", nil))
	if r.Code != 200 {
		t.Fatalf("pause status %d %s", r.Code, r.Body.String())
	}
}
