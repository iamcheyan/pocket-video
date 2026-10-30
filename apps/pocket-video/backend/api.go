package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	var version int
	_ = s.db.QueryRow("SELECT COALESCE(MAX(version),0) FROM schema_version").Scan(&version)
	account, _ := s.provider.Account(r.Context())
	mode := account.Mode
	if mode == "" {
		mode = "demo"
	}
	jsonOut(w, map[string]any{
		"ok":             true,
		"provider":       s.provider.Name(),
		"mock":           s.provider.Name() == "mock",
		"mode":           mode,
		"authenticated":  account.Authenticated,
		"database":       "sqlite",
		"schema_version": version,
	})
}

func (s *Server) account(w http.ResponseWriter, r *http.Request) {
	account, err := s.provider.Account(r.Context())
	if err != nil {
		fail(w, http.StatusServiceUnavailable, "provider_unavailable", err.Error())
		return
	}
	if account.Collections == 0 {
		_ = s.db.QueryRow("SELECT count(*) FROM favorites").Scan(&account.Collections)
	}
	jsonOut(w, account)
}

func (s *Server) favorites(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT id,title,video_count,COALESCE(last_sync_at,'') FROM favorites ORDER BY title")
	if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	defer rows.Close()
	out := []Favorite{}
	for rows.Next() {
		var x Favorite
		if err := rows.Scan(&x.ID, &x.Title, &x.VideoCount, &x.LastSyncAt); err != nil {
			fail(w, http.StatusInternalServerError, "database", err.Error())
			return
		}
		out = append(out, x)
	}
	jsonOut(w, out)
}

// videos supports search plus optional favorite filtering so the favorites page
// and the library page share one contract.
func (s *Server) videos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	favoriteID := r.URL.Query().Get("favorite_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	where := "WHERE (title LIKE ? OR creator LIKE ?)"
	args := []any{"%" + q + "%", "%" + q + "%"}
	if favoriteID != "" {
		where += " AND favorite_id = ?"
		args = append(args, favoriteID)
	}
	var total, downloaded int
	if err := s.db.QueryRow("SELECT count(*) FROM videos "+where, args...).Scan(&total); err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	rows, err := s.db.Query("SELECT id,favorite_id,title,creator,duration,COALESCE(cover,''),COALESCE(video_path,''),COALESCE(audio_path,'') FROM videos "+where+" ORDER BY id LIMIT ? OFFSET ?", append(args, limit, offset)...)
	if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	defer rows.Close()
	out := []Video{}
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.FavoriteID, &v.Title, &v.Creator, &v.Duration, &v.Cover, &v.VideoPath, &v.AudioPath); err != nil {
			fail(w, http.StatusInternalServerError, "database", err.Error())
			return
		}
		v.HasVideo = v.VideoPath != ""
		v.HasAudio = v.AudioPath != ""
		if v.Cover != "" {
			v.Cover = "/api/cover/" + v.ID
		}
		out = append(out, v)
	}
	_ = s.db.QueryRow("SELECT count(*) FROM videos WHERE COALESCE(video_path,'')<>''").Scan(&downloaded)
	jsonOut(w, map[string]any{"items": out, "limit": limit, "offset": offset, "total": total, "downloaded": downloaded})
}

func (s *Server) video(w http.ResponseWriter, r *http.Request) {
	var v Video
	err := s.db.QueryRow("SELECT id,favorite_id,title,creator,duration,COALESCE(cover,''),COALESCE(video_path,''),COALESCE(audio_path,'') FROM videos WHERE id=?", r.PathValue("id")).Scan(&v.ID, &v.FavoriteID, &v.Title, &v.Creator, &v.Duration, &v.Cover, &v.VideoPath, &v.AudioPath)
	if err == sql.ErrNoRows {
		fail(w, http.StatusNotFound, "not_found", "video not found")
		return
	}
	if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	v.HasVideo = v.VideoPath != ""
	v.HasAudio = v.AudioPath != ""
	if v.Cover != "" {
		v.Cover = "/api/cover/" + v.ID
	}
	jsonOut(w, v)
}

const taskSelect = `SELECT t.id,t.video_id,COALESCE(v.title,''),t.kind,t.status,t.downloaded_bytes,t.total_bytes,COALESCE(t.error,''),COALESCE(t.updated_at,'')
FROM download_tasks t LEFT JOIN videos v ON v.id=t.video_id`

func scanTask(scan func(...any) error) (Task, error) {
	var t Task
	err := scan(&t.ID, &t.VideoID, &t.Title, &t.Kind, &t.Status, &t.DownloadedBytes, &t.TotalBytes, &t.Error, &t.UpdatedAt)
	if t.TotalBytes > 0 {
		t.Progress = int(t.DownloadedBytes * 100 / t.TotalBytes)
	}
	return t, err
}

func (s *Server) tasks(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(taskSelect + " ORDER BY t.id DESC")
	if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	defer rows.Close()
	out := []Task{}
	for rows.Next() {
		t, err := scanTask(rows.Scan)
		if err != nil {
			fail(w, http.StatusInternalServerError, "database", err.Error())
			return
		}
		out = append(out, t)
	}
	jsonOut(w, out)
}

func (s *Server) task(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		fail(w, http.StatusBadRequest, "invalid_id", "invalid task id")
		return
	}
	t, err := scanTask(s.db.QueryRow(taskSelect+" WHERE t.id=?", id).Scan)
	if err == sql.ErrNoRows {
		fail(w, http.StatusNotFound, "not_found", "task not found")
		return
	}
	if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	jsonOut(w, t)
}

func (s *Server) download(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.VideoID == "" || (req.Kind != "video" && req.Kind != "audio") {
		fail(w, http.StatusBadRequest, "invalid_request", "video_id and kind (video|audio) are required")
		return
	}
	var exists int
	if s.db.QueryRow("SELECT count(*) FROM videos WHERE id=?", req.VideoID).Scan(&exists) != nil || exists == 0 {
		fail(w, http.StatusNotFound, "not_found", "video not found")
		return
	}
	// One active job per video and media kind keeps the queue honest: a duplicate
	// click resumes a paused task or reports the existing job.
	var existingID int64
	var existingStatus string
	err := s.db.QueryRow("SELECT id,status FROM download_tasks WHERE video_id=? AND kind=? AND status IN ('queued','running','paused') ORDER BY id DESC LIMIT 1", req.VideoID, req.Kind).Scan(&existingID, &existingStatus)
	if err == nil {
		if existingStatus == "paused" {
			s.start(existingID, req)
			existingStatus = "running"
		}
		jsonOut(w, map[string]any{"id": existingID, "status": existingStatus, "existing": true})
		return
	}
	res, err := s.db.Exec("INSERT INTO download_tasks(video_id,kind,status) VALUES(?,?,'queued')", req.VideoID, req.Kind)
	if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	s.start(id, req)
	w.WriteHeader(http.StatusAccepted)
	jsonOut(w, map[string]any{"id": id, "status": "queued"})
}

func (s *Server) pause(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		fail(w, http.StatusBadRequest, "invalid_id", "invalid task id")
		return
	}
	s.mu.Lock()
	cancel := s.cancels[id]
	state := s.workerStates[id]
	s.mu.Unlock()
	if cancel == nil || state != "running" {
		fail(w, http.StatusConflict, "invalid_state", "task is not running")
		return
	}
	cancel()
	for {
		s.mu.Lock()
		state = s.workerStates[id]
		s.mu.Unlock()
		if state == "paused" || state == "failed" || state == "done" {
			break
		}
		time.Sleep(time.Millisecond)
	}
	var status string
	_ = s.db.QueryRow("SELECT status FROM download_tasks WHERE id=?", id).Scan(&status)
	if status != "paused" {
		fail(w, http.StatusConflict, "invalid_state", "task finished before pause completed")
		return
	}
	jsonOut(w, map[string]string{"status": "paused"})
}

func (s *Server) resume(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		fail(w, http.StatusBadRequest, "invalid_id", "invalid task id")
		return
	}
	var req DownloadRequest
	var status string
	if err = s.db.QueryRow("SELECT video_id,kind,status FROM download_tasks WHERE id=?", id).Scan(&req.VideoID, &req.Kind, &status); err == sql.ErrNoRows {
		fail(w, http.StatusNotFound, "not_found", "task not found")
		return
	} else if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	if status != "paused" && status != "queued" {
		fail(w, http.StatusConflict, "invalid_state", "only queued or paused tasks can resume")
		return
	}
	s.start(id, req)
	jsonOut(w, map[string]string{"status": "running"})
}

// deleteTask removes a finished or idle queue entry. Completed media stays in
// the library; partial files are cleaned up so a retry starts clean.
func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		fail(w, http.StatusBadRequest, "invalid_id", "invalid task id")
		return
	}
	var videoID, kind, status string
	if err = s.db.QueryRow("SELECT video_id,kind,status FROM download_tasks WHERE id=?", id).Scan(&videoID, &kind, &status); err == sql.ErrNoRows {
		fail(w, http.StatusNotFound, "not_found", "task not found")
		return
	} else if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	if status == "running" {
		fail(w, http.StatusConflict, "invalid_state", "pause a running task before removing it")
		return
	}
	if status != "done" {
		s.cleanupPartial(videoID, kind)
	}
	if _, err = s.db.Exec("DELETE FROM download_tasks WHERE id=?", id); err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) cleanupPartial(videoID, kind string) {
	var dir string
	if err := s.db.QueryRow("SELECT download_dir FROM settings WHERE id=1").Scan(&dir); err != nil || dir == "" {
		dir = s.providerConfig.OutputDir
	}
	base := filepath.Join(dir, videoID+"."+kind)
	for _, candidate := range []string{base + ".mp4", base + ".m4a"} {
		if rel, err := filepath.Rel(dir, candidate); err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		_ = os.Remove(candidate)
		_ = os.Remove(candidate + ".source" + filepath.Ext(candidate))
	}
}

func (s *Server) sync(w http.ResponseWriter, r *http.Request) {
	if err := s.syncFavorites(r.Context()); err != nil {
		fail(w, http.StatusServiceUnavailable, "provider_unavailable", err.Error())
		return
	}
	var favorites, videos int
	_ = s.db.QueryRow("SELECT (SELECT count(*) FROM favorites),(SELECT count(*) FROM videos)").Scan(&favorites, &videos)
	jsonOut(w, map[string]any{"favorites": favorites, "videos": videos, "synced_at": time.Now().UTC().Format(time.RFC3339)})
}

func (s *Server) settings(w http.ResponseWriter, r *http.Request) {
	var x Settings
	var auto int
	err := s.db.QueryRow("SELECT language,theme,cache_dir,download_dir,sync_minutes,auto_sync,proxy FROM settings WHERE id=1").Scan(&x.Language, &x.Theme, &x.CacheDir, &x.DownloadDir, &x.SyncMinutes, &auto, &x.Proxy)
	if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	x.AutoSync = auto != 0
	jsonOut(w, x)
}

func (s *Server) updateSettings(w http.ResponseWriter, r *http.Request) {
	var x Settings
	if json.NewDecoder(r.Body).Decode(&x) != nil {
		fail(w, http.StatusBadRequest, "invalid_request", "invalid settings JSON")
		return
	}
	if !map[string]bool{"en": true, "zh-CN": true, "ja": true}[x.Language] ||
		!map[string]bool{"system": true, "light": true, "dark": true}[x.Theme] ||
		x.SyncMinutes < 1 || x.SyncMinutes > 10080 ||
		!filepath.IsAbs(x.CacheDir) || !filepath.IsAbs(x.DownloadDir) {
		fail(w, http.StatusBadRequest, "invalid_settings", "language/theme unsupported, sync_minutes must be 1-10080, and cache/download directories must be absolute")
		return
	}
	auto := 0
	if x.AutoSync {
		auto = 1
	}
	_, err := s.db.Exec("INSERT INTO settings(id,language,theme,cache_dir,download_dir,sync_minutes,auto_sync,proxy) VALUES(1,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET language=excluded.language,theme=excluded.theme,cache_dir=excluded.cache_dir,download_dir=excluded.download_dir,sync_minutes=excluded.sync_minutes,auto_sync=excluded.auto_sync,proxy=excluded.proxy", x.Language, x.Theme, x.CacheDir, x.DownloadDir, x.SyncMinutes, auto, x.Proxy)
	if err != nil {
		fail(w, http.StatusInternalServerError, "database", err.Error())
		return
	}
	jsonOut(w, x)
}

func (s *Server) media(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	kind := r.URL.Query().Get("kind")
	column := "video_path"
	if kind == "audio" {
		column = "audio_path"
	} else if kind != "" && kind != "video" {
		fail(w, http.StatusBadRequest, "invalid_kind", "kind must be video or audio")
		return
	}
	var path string
	err := s.db.QueryRow("SELECT COALESCE("+column+",'') FROM videos WHERE id=?", id).Scan(&path)
	if err != nil || path == "" {
		fail(w, http.StatusNotFound, "media_not_found", "downloaded media is unavailable")
		return
	}
	http.ServeFile(w, r, path)
}

// cover serves the cached artwork extracted when a video finished downloading.
func (s *Server) cover(w http.ResponseWriter, r *http.Request) {
	var path string
	if err := s.db.QueryRow("SELECT COALESCE(cover,'') FROM videos WHERE id=?", r.PathValue("id")).Scan(&path); err != nil || path == "" {
		fail(w, http.StatusNotFound, "cover_not_found", "no cached artwork for this video")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, path)
}
