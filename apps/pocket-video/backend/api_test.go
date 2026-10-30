package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestFavoritesFilterAndAccount(t *testing.T) {
	s := testServer(t)
	m := http.NewServeMux()
	s.routes(m)

	var favorites []Favorite
	if err := json.Unmarshal(request(m, "GET", "/api/favorites", "").Body.Bytes(), &favorites); err != nil {
		t.Fatal(err)
	}
	if len(favorites) != len(demoFavorites) {
		t.Fatalf("favorites=%d, want %d", len(favorites), len(demoFavorites))
	}
	total := 0
	for _, f := range favorites {
		if f.LastSyncAt == "" || f.VideoCount == 0 {
			t.Fatalf("favorite %q missing sync metadata: %+v", f.ID, f)
		}
		total += f.VideoCount
	}
	if total != len(demoVideos) {
		t.Fatalf("favorite video counts sum to %d, want %d", total, len(demoVideos))
	}

	var page struct {
		Items []Video `json:"items"`
		Total int     `json:"total"`
	}
	if err := json.Unmarshal(request(m, "GET", "/api/videos?favorite_id=fav-audio", "").Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 3 || len(page.Items) != 3 {
		t.Fatalf("favorite filter total=%d items=%d", page.Total, len(page.Items))
	}
	for _, v := range page.Items {
		if v.FavoriteID != "fav-audio" {
			t.Fatalf("filter leaked %q from %q", v.ID, v.FavoriteID)
		}
	}

	var search struct {
		Items []Video `json:"items"`
		Total int     `json:"total"`
	}
	if err := json.Unmarshal(request(m, "GET", "/api/videos?q=piano", "").Body.Bytes(), &search); err != nil {
		t.Fatal(err)
	}
	if search.Total != 2 {
		t.Fatalf("search piano total=%d, want 2", search.Total)
	}

	var account Account
	if err := json.Unmarshal(request(m, "GET", "/api/account", "").Body.Bytes(), &account); err != nil {
		t.Fatal(err)
	}
	if account.Mode != "demo" || account.Provider != "mock" || !account.Authenticated || account.Collections != len(demoFavorites) {
		t.Fatalf("unexpected account payload: %+v", account)
	}
}

func TestDuplicateDownloadIsIdempotent(t *testing.T) {
	s := testServer(t)
	m := http.NewServeMux()
	s.routes(m)
	first := request(m, "POST", "/api/download", `{"video_id":"BV1DEMO001","kind":"video"}`)
	if first.Code != http.StatusAccepted {
		t.Fatalf("first download: %d %s", first.Code, first.Body.String())
	}
	second := request(m, "POST", "/api/download", `{"video_id":"BV1DEMO001","kind":"video"}`)
	if second.Code != http.StatusOK {
		t.Fatalf("second download: %d %s", second.Code, second.Body.String())
	}
	var firstID, secondID int64
	_ = json.Unmarshal(first.Body.Bytes(), &struct {
		ID *int64 `json:"id"`
	}{&firstID})
	var existing struct {
		ID       int64 `json:"id"`
		Existing bool  `json:"existing"`
	}
	if err := json.Unmarshal(second.Body.Bytes(), &existing); err != nil {
		t.Fatal(err)
	}
	secondID = existing.ID
	if !existing.Existing || firstID != secondID {
		t.Fatalf("duplicate was not collapsed: first=%d second=%+v", firstID, existing)
	}
	var tasks int
	if err := s.db.QueryRow("SELECT count(*) FROM download_tasks").Scan(&tasks); err != nil || tasks != 1 {
		t.Fatalf("tasks=%d err=%v", tasks, err)
	}
}

func TestDeleteTaskRulesAndPartialCleanup(t *testing.T) {
	s := testServer(t)
	m := http.NewServeMux()
	s.routes(m)
	if r := request(m, "POST", "/api/download", `{"video_id":"BV1DEMO001","kind":"video"}`); r.Code != http.StatusAccepted {
		t.Fatalf("download: %d %s", r.Code, r.Body.String())
	}
	var id int64
	if err := s.db.QueryRow("SELECT id FROM download_tasks").Scan(&id); err != nil {
		t.Fatal(err)
	}
	var status string
	var downloaded int64
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		_ = s.db.QueryRow("SELECT status,downloaded_bytes FROM download_tasks WHERE id=?", id).Scan(&status, &downloaded)
		if downloaded > 0 || status == "failed" || status == "done" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if downloaded == 0 {
		t.Fatalf("download never progressed: status=%s", status)
	}
	if r := request(m, "POST", "/api/task/"+strconv.FormatInt(id, 10)+"/pause", ""); r.Code != http.StatusOK {
		t.Fatalf("pause: %d %s", r.Code, r.Body.String())
	}
	waitTask(t, s, id, "paused")
	partial := filepath.Join(s.root, "downloads", "BV1DEMO001.video.mp4")
	if st, err := os.Stat(partial); err != nil || st.Size() == 0 {
		t.Fatalf("partial media missing before delete: %v", err)
	}
	if r := request(m, "DELETE", "/api/task/"+strconv.FormatInt(id, 10), ""); r.Code != http.StatusNoContent {
		t.Fatalf("delete paused: %d %s", r.Code, r.Body.String())
	}
	var tasks int
	if err := s.db.QueryRow("SELECT count(*) FROM download_tasks").Scan(&tasks); err != nil || tasks != 0 {
		t.Fatalf("tasks after delete=%d err=%v", tasks, err)
	}
	if _, err := os.Stat(partial); !os.IsNotExist(err) {
		t.Fatalf("partial media was not cleaned up: %v", err)
	}

	// A running task must be paused first.
	if _, err := s.db.Exec("INSERT INTO download_tasks(video_id,kind,status) VALUES('BV1DEMO002','video','running')"); err != nil {
		t.Fatal(err)
	}
	var runningID int64
	_ = s.db.QueryRow("SELECT id FROM download_tasks").Scan(&runningID)
	if r := request(m, "DELETE", "/api/task/"+strconv.FormatInt(runningID, 10), ""); r.Code != http.StatusConflict {
		t.Fatalf("delete running: %d %s", r.Code, r.Body.String())
	}
}

func TestSettingsValidationRejectsRelativeDirectories(t *testing.T) {
	s := testServer(t)
	m := http.NewServeMux()
	s.routes(m)
	cases := []string{
		`{"language":"en","theme":"dark","cache_dir":"cache","download_dir":"/media","sync_minutes":30,"auto_sync":false,"proxy":""}`,
		`{"language":"fr","theme":"dark","cache_dir":"/cache","download_dir":"/media","sync_minutes":30,"auto_sync":false,"proxy":""}`,
		`{"language":"en","theme":"dark","cache_dir":"/cache","download_dir":"/media","sync_minutes":0,"auto_sync":false,"proxy":""}`,
	}
	for _, body := range cases {
		if r := request(m, "PUT", "/api/settings", body); r.Code != http.StatusBadRequest {
			t.Fatalf("invalid settings accepted: %d %s", r.Code, body)
		}
	}
	var lang string
	_ = s.db.QueryRow("SELECT language FROM settings WHERE id=1").Scan(&lang)
	if lang != "en" {
		t.Fatalf("rejected settings mutated the row: %q", lang)
	}
}

func TestDownloadedVideoGetsServedCover(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg is required for demo artwork")
	}
	s := testServer(t)
	m := http.NewServeMux()
	s.routes(m)
	if r := request(m, "POST", "/api/download", `{"video_id":"BV1DEMO003","kind":"video"}`); r.Code != http.StatusAccepted {
		t.Fatalf("download: %d %s", r.Code, r.Body.String())
	}
	var id int64
	if err := s.db.QueryRow("SELECT id FROM download_tasks").Scan(&id); err != nil {
		t.Fatal(err)
	}
	waitTask(t, s, id, "done")
	var coverPath string
	if err := s.db.QueryRow("SELECT COALESCE(cover,'') FROM videos WHERE id='BV1DEMO003'").Scan(&coverPath); err != nil || coverPath == "" {
		t.Fatalf("cover path=%q err=%v", coverPath, err)
	}
	if st, err := os.Stat(coverPath); err != nil || st.Size() == 0 {
		t.Fatalf("cover file missing: %v", err)
	}
	var page struct {
		Items []Video `json:"items"`
	}
	if err := json.Unmarshal(request(m, "GET", "/api/videos?q=PocketJS", "").Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	for _, video := range page.Items {
		if video.ID == "BV1DEMO003" {
			if video.Cover != "/api/cover/BV1DEMO003" {
				t.Fatalf("API cover should be a URL, got %q", video.Cover)
			}
			if video.VideoPath != "" {
				t.Fatalf("local media path leaked: %q", video.VideoPath)
			}
		}
	}
	cover := request(m, "GET", "/api/cover/BV1DEMO003", "")
	if cover.Code != http.StatusOK || cover.Body.Len() == 0 || cover.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("cover response status=%d type=%q bytes=%d", cover.Code, cover.Header().Get("Content-Type"), cover.Body.Len())
	}
}
