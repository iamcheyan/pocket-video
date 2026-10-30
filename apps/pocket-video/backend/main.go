package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Favorite struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	VideoCount int    `json:"video_count"`
	LastSyncAt string `json:"last_sync_at,omitempty"`
}
type Video struct {
	ID         string `json:"id"`
	FavoriteID string `json:"favorite_id"`
	Title      string `json:"title"`
	Creator    string `json:"creator"`
	Duration   int    `json:"duration"`
	Cover      string `json:"cover"`
	HasVideo   bool   `json:"has_video"`
	HasAudio   bool   `json:"has_audio"`
	VideoPath  string `json:"video_path,omitempty"`
	AudioPath  string `json:"audio_path,omitempty"`
}
type Task struct {
	ID              int64  `json:"id"`
	VideoID         string `json:"video_id"`
	Kind            string `json:"kind"`
	Status          string `json:"status"`
	DownloadedBytes int64  `json:"downloaded_bytes"`
	TotalBytes      int64  `json:"total_bytes"`
	Error           string `json:"error,omitempty"`
	Progress        int    `json:"progress"`
}
type Settings struct {
	Language    string `json:"language"`
	Theme       string `json:"theme"`
	CacheDir    string `json:"cache_dir"`
	DownloadDir string `json:"download_dir"`
	SyncMinutes int    `json:"sync_minutes"`
	AutoSync    bool   `json:"auto_sync"`
	Proxy       string `json:"proxy"`
}
type Query struct{ Q string }
type DownloadRequest struct {
	VideoID string `json:"video_id"`
	Kind    string `json:"kind"`
}
type Progress struct{ Downloaded, Total int64 }
type Result struct{ Path string }
type VideoProvider interface {
	Name() string
	SyncFavorites(context.Context) ([]Favorite, error)
	ListVideos(context.Context, Query) ([]Video, error)
	Download(context.Context, DownloadRequest, func(Progress)) (Result, error)
}
type MockProvider struct{}

func (MockProvider) Name() string { return "mock" }
func (MockProvider) SyncFavorites(context.Context) ([]Favorite, error) {
	return []Favorite{{"fav-1", "PocketJS Favorites", 3, time.Now().UTC().Format(time.RFC3339)}}, nil
}
func (MockProvider) ListVideos(_ context.Context, q Query) ([]Video, error) {
	all := []Video{{"BV1DEMO001", "fav-1", "PocketVideo Architecture Demo", "PocketVideo Lab", 305, "", false, false, "", ""}, {"BV1DEMO002", "fav-1", "Offline First: A Practical Guide", "PocketVideo Lab", 428, "", false, false, "", ""}, {"BV1DEMO003", "fav-1", "PocketJS UI on Small Screens", "PocketVideo Lab", 247, "", false, false, "", ""}}
	if q.Q == "" {
		return all, nil
	}
	out := []Video{}
	for _, v := range all {
		if strings.Contains(strings.ToLower(v.Title+" "+v.Creator), strings.ToLower(q.Q)) {
			out = append(out, v)
		}
	}
	return out, nil
}
func (MockProvider) Download(ctx context.Context, r DownloadRequest, p func(Progress)) (Result, error) {
	const total int64 = 5 * 1024 * 1024
	for n := int64(0); n <= total; n += total / 20 {
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		default:
		}
		p(Progress{n, total})
		time.Sleep(200 * time.Millisecond)
	}
	return Result{}, nil
}

type YtDlpProvider struct{ Binary string }

func (y YtDlpProvider) Name() string { return "yt-dlp" }
func (y YtDlpProvider) SyncFavorites(context.Context) ([]Favorite, error) {
	return nil, errors.New("Bilibili login is not configured; yt-dlp provider cannot sync favorites")
}
func (y YtDlpProvider) ListVideos(context.Context, Query) ([]Video, error) {
	return nil, errors.New("Bilibili login is not configured; yt-dlp provider cannot list favorites")
}
func (y YtDlpProvider) Download(ctx context.Context, r DownloadRequest, p func(Progress)) (Result, error) {
	bin := y.Binary
	if bin == "" {
		bin = "yt-dlp"
	}
	cmd := exec.CommandContext(ctx, bin, "--no-progress", "--newline", "-o", "%(id)s.%(ext)s", "https://www.bilibili.com/video/"+r.VideoID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{}, fmt.Errorf("yt-dlp: %w: %s", err, strings.TrimSpace(string(out)))
	}
	p(Progress{})
	return Result{}, nil
}

type Server struct {
	db       *sql.DB
	provider VideoProvider
	root     string
	mu       sync.Mutex
	cancels  map[int64]context.CancelFunc
}

func main() {
	root := os.Getenv("POCKETVIDEO_DATA")
	if root == "" {
		root = "./data"
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(root, "pocketvideo.db"))
	if err != nil {
		log.Fatal(err)
	}
	if err = migrate(db); err != nil {
		log.Fatal(err)
	}
	var p VideoProvider = MockProvider{}
	if os.Getenv("POCKETVIDEO_PROVIDER") == "ytdlp" {
		p = YtDlpProvider{}
	}
	s := &Server{db: db, provider: p, root: root, cancels: map[int64]context.CancelFunc{}}
	s.seed()
	mux := http.NewServeMux()
	s.routes(mux)
	addr := os.Getenv("POCKETVIDEO_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("PocketVideo %s listening on %s", p.Name(), addr)
	log.Fatal(http.ListenAndServe(addr, cors(mux)))
}
func migrate(db *sql.DB) error {
	_, e := db.Exec(`PRAGMA journal_mode=WAL; CREATE TABLE IF NOT EXISTS schema_version(version INTEGER NOT NULL); INSERT INTO schema_version SELECT 1 WHERE NOT EXISTS(SELECT 1 FROM schema_version); CREATE TABLE IF NOT EXISTS favorites(id TEXT PRIMARY KEY,title TEXT NOT NULL,video_count INTEGER NOT NULL DEFAULT 0,last_sync_at TEXT); CREATE TABLE IF NOT EXISTS videos(id TEXT PRIMARY KEY,favorite_id TEXT NOT NULL,title TEXT NOT NULL,creator TEXT NOT NULL,duration INTEGER NOT NULL,cover TEXT,video_path TEXT,audio_path TEXT,FOREIGN KEY(favorite_id) REFERENCES favorites(id)); CREATE INDEX IF NOT EXISTS videos_title_idx ON videos(title); CREATE TABLE IF NOT EXISTS download_tasks(id INTEGER PRIMARY KEY AUTOINCREMENT,video_id TEXT NOT NULL,kind TEXT NOT NULL CHECK(kind IN ('video','audio')),status TEXT NOT NULL CHECK(status IN ('queued','running','paused','done','failed')),downloaded_bytes INTEGER NOT NULL DEFAULT 0,total_bytes INTEGER NOT NULL DEFAULT 0,error TEXT,updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP); CREATE TABLE IF NOT EXISTS settings(id INTEGER PRIMARY KEY CHECK(id=1),language TEXT NOT NULL,theme TEXT NOT NULL,cache_dir TEXT NOT NULL,download_dir TEXT NOT NULL,sync_minutes INTEGER NOT NULL,auto_sync INTEGER NOT NULL,proxy TEXT NOT NULL);`)
	return e
}
func (s *Server) seed() {
	var n int
	s.db.QueryRow("SELECT count(*) FROM favorites").Scan(&n)
	if n > 0 {
		return
	}
	f, _ := (MockProvider{}).SyncFavorites(context.Background())
	for _, x := range f {
		s.db.Exec("INSERT INTO favorites(id,title,video_count,last_sync_at) VALUES(?,?,?,?)", x.ID, x.Title, x.VideoCount, x.LastSyncAt)
	}
	vs, _ := (MockProvider{}).ListVideos(context.Background(), Query{})
	for _, v := range vs {
		s.db.Exec("INSERT INTO videos(id,favorite_id,title,creator,duration,cover) VALUES(?,?,?,?,?,?)", v.ID, v.FavoriteID, v.Title, v.Creator, v.Duration, v.Cover)
	}
	s.db.Exec("INSERT OR IGNORE INTO settings VALUES(1,'en','system',?,?,60,0,'')", filepath.Join(s.root, "cache"), filepath.Join(s.root, "downloads"))
}
func (s *Server) routes(m *http.ServeMux) {
	m.HandleFunc("GET /api/health", s.health)
	m.HandleFunc("GET /api/favorites", s.favorites)
	m.HandleFunc("GET /api/videos", s.videos)
	m.HandleFunc("GET /api/videos/{id}", s.video)
	m.HandleFunc("GET /api/task", s.tasks)
	m.HandleFunc("POST /api/download", s.download)
	m.HandleFunc("POST /api/task/{id}/pause", s.pause)
	m.HandleFunc("POST /api/task/{id}/resume", s.resume)
	m.HandleFunc("POST /api/sync", s.sync)
	m.HandleFunc("GET /api/settings", s.settings)
	m.HandleFunc("PUT /api/settings", s.updateSettings)
	m.HandleFunc("GET /api/media/{id}", s.media)
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Range")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func jsonOut(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, code, msg string) {
	w.WriteHeader(status)
	jsonOut(w, map[string]any{"error": map[string]string{"code": code, "message": msg}})
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	var version int
	s.db.QueryRow("SELECT version FROM schema_version LIMIT 1").Scan(&version)
	jsonOut(w, map[string]any{"ok": true, "provider": s.provider.Name(), "mock": s.provider.Name() == "mock", "database": "sqlite", "schema_version": version})
}
func (s *Server) favorites(w http.ResponseWriter, r *http.Request) {
	rows, e := s.db.Query("SELECT id,title,video_count,COALESCE(last_sync_at,'') FROM favorites ORDER BY title")
	if e != nil {
		fail(w, 500, "database", e.Error())
		return
	}
	defer rows.Close()
	out := []Favorite{}
	for rows.Next() {
		var x Favorite
		rows.Scan(&x.ID, &x.Title, &x.VideoCount, &x.LastSyncAt)
		out = append(out, x)
	}
	jsonOut(w, out)
}
func (s *Server) videos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, e := s.db.Query("SELECT id,favorite_id,title,creator,duration,COALESCE(cover,''),video_path,audio_path FROM videos WHERE title LIKE ? OR creator LIKE ? ORDER BY title LIMIT ? OFFSET ?", "%"+q+"%", "%"+q+"%", limit, offset)
	if e != nil {
		fail(w, 500, "database", e.Error())
		return
	}
	defer rows.Close()
	out := []Video{}
	for rows.Next() {
		var v Video
		rows.Scan(&v.ID, &v.FavoriteID, &v.Title, &v.Creator, &v.Duration, &v.Cover, &v.VideoPath, &v.AudioPath)
		v.HasVideo = v.VideoPath != ""
		v.HasAudio = v.AudioPath != ""
		out = append(out, v)
	}
	jsonOut(w, map[string]any{"items": out, "limit": limit, "offset": offset})
}
func (s *Server) video(w http.ResponseWriter, r *http.Request) {
	var v Video
	e := s.db.QueryRow("SELECT id,favorite_id,title,creator,duration,COALESCE(cover,''),COALESCE(video_path,''),COALESCE(audio_path,'') FROM videos WHERE id=?", r.PathValue("id")).Scan(&v.ID, &v.FavoriteID, &v.Title, &v.Creator, &v.Duration, &v.Cover, &v.VideoPath, &v.AudioPath)
	if e == sql.ErrNoRows {
		fail(w, 404, "not_found", "video not found")
		return
	}
	if e != nil {
		fail(w, 500, "database", e.Error())
		return
	}
	v.HasVideo = v.VideoPath != ""
	v.HasAudio = v.AudioPath != ""
	jsonOut(w, v)
}
func (s *Server) tasks(w http.ResponseWriter, r *http.Request) {
	rows, e := s.db.Query("SELECT id,video_id,kind,status,downloaded_bytes,total_bytes,COALESCE(error,'') FROM download_tasks ORDER BY id DESC")
	if e != nil {
		fail(w, 500, "database", e.Error())
		return
	}
	defer rows.Close()
	out := []Task{}
	for rows.Next() {
		var t Task
		rows.Scan(&t.ID, &t.VideoID, &t.Kind, &t.Status, &t.DownloadedBytes, &t.TotalBytes, &t.Error)
		if t.TotalBytes > 0 {
			t.Progress = int(t.DownloadedBytes * 100 / t.TotalBytes)
		}
		out = append(out, t)
	}
	jsonOut(w, out)
}
func (s *Server) download(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.VideoID == "" || (req.Kind != "video" && req.Kind != "audio") {
		fail(w, 400, "invalid_request", "video_id and kind (video|audio) are required")
		return
	}
	var exists int
	if s.db.QueryRow("SELECT count(*) FROM videos WHERE id=?", req.VideoID).Scan(&exists) != nil || exists == 0 {
		fail(w, 404, "not_found", "video not found")
		return
	}
	res, e := s.db.Exec("INSERT INTO download_tasks(video_id,kind,status,total_bytes) VALUES(?,?,'queued',5242880)", req.VideoID, req.Kind)
	if e != nil {
		fail(w, 500, "database", e.Error())
		return
	}
	id, _ := res.LastInsertId()
	s.start(id, req)
	w.WriteHeader(202)
	jsonOut(w, map[string]any{"id": id, "status": "queued"})
}
func (s *Server) start(id int64, req DownloadRequest) {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancels[id] = cancel
	s.mu.Unlock()
	go func() {
		defer func() { s.mu.Lock(); delete(s.cancels, id); s.mu.Unlock() }()
		s.db.Exec("UPDATE download_tasks SET status='running',error=NULL,updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
		result, err := (MockProvider{}).Download(ctx, req, func(p Progress) {
			s.db.Exec("UPDATE download_tasks SET downloaded_bytes=?,total_bytes=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", p.Downloaded, p.Total, id)
		})
		if ctx.Err() != nil {
			s.db.Exec("UPDATE download_tasks SET status='paused' WHERE id=?", id)
			return
		}
		if err != nil {
			s.db.Exec("UPDATE download_tasks SET status='failed',error=? WHERE id=?", err.Error(), id)
			return
		}
		os.MkdirAll(filepath.Join(s.root, "downloads"), 0700)
		name := req.VideoID + "." + req.Kind
		path := filepath.Join(s.root, "downloads", name)
		os.WriteFile(path, []byte("PocketVideo offline demo media: "+req.VideoID), 0600)
		_ = result
		s.db.Exec("UPDATE download_tasks SET status='done',downloaded_bytes=total_bytes WHERE id=?", id)
		column := "video_path"
		if req.Kind == "audio" {
			column = "audio_path"
		}
		s.db.Exec("UPDATE videos SET "+column+"=? WHERE id=?", path, req.VideoID)
	}()
}
func (s *Server) pause(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		fail(w, 400, "invalid_id", "invalid task id")
		return
	}
	s.mu.Lock()
	c := s.cancels[id]
	s.mu.Unlock()
	if c == nil {
		fail(w, 409, "invalid_state", "task is not running")
		return
	}
	c()
	jsonOut(w, map[string]string{"status": "pausing"})
}
func (s *Server) resume(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		fail(w, 400, "invalid_id", "invalid task id")
		return
	}
	var req DownloadRequest
	var status string
	e = s.db.QueryRow("SELECT video_id,kind,status FROM download_tasks WHERE id=?", id).Scan(&req.VideoID, &req.Kind, &status)
	if e == sql.ErrNoRows {
		fail(w, 404, "not_found", "task not found")
		return
	}
	if status != "paused" {
		fail(w, 409, "invalid_state", "only paused tasks can resume")
		return
	}
	s.start(id, req)
	jsonOut(w, map[string]string{"status": "running"})
}
func (s *Server) sync(w http.ResponseWriter, r *http.Request) {
	fs, e := s.provider.SyncFavorites(r.Context())
	if e != nil {
		fail(w, 503, "provider_unavailable", e.Error())
		return
	}
	for _, f := range fs {
		s.db.Exec("INSERT INTO favorites(id,title,video_count,last_sync_at) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET title=excluded.title,video_count=excluded.video_count,last_sync_at=excluded.last_sync_at", f.ID, f.Title, f.VideoCount, f.LastSyncAt)
	}
	vs, e := s.provider.ListVideos(r.Context(), Query{})
	if e != nil {
		fail(w, 503, "provider_unavailable", e.Error())
		return
	}
	for _, v := range vs {
		s.db.Exec("INSERT INTO videos(id,favorite_id,title,creator,duration,cover) VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET title=excluded.title", v.ID, v.FavoriteID, v.Title, v.Creator, v.Duration, v.Cover)
	}
	jsonOut(w, map[string]any{"favorites": len(fs), "videos": len(vs)})
}
func (s *Server) settings(w http.ResponseWriter, r *http.Request) {
	var x Settings
	var auto int
	e := s.db.QueryRow("SELECT language,theme,cache_dir,download_dir,sync_minutes,auto_sync,proxy FROM settings WHERE id=1").Scan(&x.Language, &x.Theme, &x.CacheDir, &x.DownloadDir, &x.SyncMinutes, &auto, &x.Proxy)
	if e != nil {
		fail(w, 500, "database", e.Error())
		return
	}
	x.AutoSync = auto != 0
	jsonOut(w, x)
}
func (s *Server) updateSettings(w http.ResponseWriter, r *http.Request) {
	var x Settings
	if json.NewDecoder(r.Body).Decode(&x) != nil {
		fail(w, 400, "invalid_request", "invalid settings JSON")
		return
	}
	if !map[string]bool{"en": true, "zh-CN": true, "ja": true}[x.Language] || !map[string]bool{"system": true, "light": true, "dark": true}[x.Theme] || x.SyncMinutes < 1 {
		fail(w, 400, "invalid_settings", "unsupported language/theme or sync_minutes < 1")
		return
	}
	auto := 0
	if x.AutoSync {
		auto = 1
	}
	_, e := s.db.Exec("INSERT INTO settings(id,language,theme,cache_dir,download_dir,sync_minutes,auto_sync,proxy) VALUES(1,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET language=excluded.language,theme=excluded.theme,cache_dir=excluded.cache_dir,download_dir=excluded.download_dir,sync_minutes=excluded.sync_minutes,auto_sync=excluded.auto_sync,proxy=excluded.proxy", x.Language, x.Theme, x.CacheDir, x.DownloadDir, x.SyncMinutes, auto, x.Proxy)
	if e != nil {
		fail(w, 500, "database", e.Error())
		return
	}
	jsonOut(w, x)
}
func (s *Server) media(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var path string
	e := s.db.QueryRow("SELECT video_path FROM videos WHERE id=?", id).Scan(&path)
	if e != nil || path == "" {
		fail(w, 404, "media_not_found", "downloaded video is unavailable")
		return
	}
	http.ServeFile(w, r, path)
}
