package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
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
	Download(context.Context, DownloadRequest, string, int64, func(Progress)) (Result, error)
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

type Server struct {
	db             *sql.DB
	provider       VideoProvider
	root           string
	mu             sync.Mutex
	cancels        map[int64]context.CancelFunc
	providerConfig ProviderDownloadConfig
	workerStates   map[int64]string
	autoSyncEvery  time.Duration
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
	cfg := ProviderDownloadConfig{OutputDir: filepath.Join(root, "downloads")}
	if os.Getenv("POCKETVIDEO_PROVIDER") == "ytdlp" {
		cfg.Proxy = os.Getenv("POCKETVIDEO_PROXY")
		p = YtDlpProvider{Binary: os.Getenv("YT_DLP_BINARY"), OutputDir: cfg.OutputDir, Proxy: cfg.Proxy}
	}
	s := &Server{db: db, provider: p, root: root, cancels: map[int64]context.CancelFunc{}, providerConfig: cfg, workerStates: map[int64]string{}}
	s.seed()
	if _, err := db.Exec("UPDATE download_tasks SET status='paused',error='service restarted; resume to continue' WHERE status='running'"); err != nil {
		log.Fatal(err)
	}
	s.startAutoSync(context.Background())
	mux := http.NewServeMux()
	s.routes(mux)
	addr := os.Getenv("POCKETVIDEO_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("PocketVideo %s listening on %s", p.Name(), addr)
	log.Fatal(http.ListenAndServe(addr, cors(mux)))
}
func (s *Server) seed() {
	var n int
	s.db.QueryRow("SELECT count(*) FROM favorites").Scan(&n)
	if n == 0 {
		f, _ := (MockProvider{}).SyncFavorites(context.Background())
		for _, x := range f {
			_, _ = s.db.Exec("INSERT OR IGNORE INTO favorites(id,title,video_count,last_sync_at) VALUES(?,?,?,?)", x.ID, x.Title, x.VideoCount, x.LastSyncAt)
		}
		vs, _ := (MockProvider{}).ListVideos(context.Background(), Query{})
		for _, v := range vs {
			_, _ = s.db.Exec("INSERT OR IGNORE INTO videos(id,favorite_id,title,creator,duration,cover) VALUES(?,?,?,?,?,?)", v.ID, v.FavoriteID, v.Title, v.Creator, v.Duration, v.Cover)
		}
	}
	_, _ = s.db.Exec("INSERT OR IGNORE INTO settings(id,language,theme,cache_dir,download_dir,sync_minutes,auto_sync,proxy) VALUES(1,'en','system',?,?,60,0,'')", filepath.Join(s.root, "cache"), filepath.Join(s.root, "downloads"))
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
	res, err := s.db.Exec("INSERT INTO download_tasks(video_id,kind,status) VALUES(?,?,'queued')", req.VideoID, req.Kind)
	if err != nil {
		fail(w, 500, "database", err.Error())
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		fail(w, 500, "database", err.Error())
		return
	}
	s.start(id, req)
	w.WriteHeader(http.StatusAccepted)
	jsonOut(w, map[string]any{"id": id, "status": "queued"})
}
func (s *Server) start(id int64, req DownloadRequest) {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	if s.cancels[id] != nil {
		s.mu.Unlock()
		cancel()
		return
	}
	s.cancels[id] = cancel
	s.workerStates[id] = "running"
	s.mu.Unlock()
	go func() {
		defer func() {
			cancel()
			s.mu.Lock()
			delete(s.cancels, id)
			if s.workerStates[id] == "running" {
				s.workerStates[id] = "done"
			}
			s.mu.Unlock()
		}()
		var offset int64
		if err := s.db.QueryRow("SELECT downloaded_bytes FROM download_tasks WHERE id=?", id).Scan(&offset); err != nil {
			return
		}
		if _, err := s.db.Exec("UPDATE download_tasks SET status='running',error=NULL,updated_at=CURRENT_TIMESTAMP WHERE id=?", id); err != nil {
			return
		}
		var downloadDir, proxy, cacheDir string
		if err := s.db.QueryRow("SELECT download_dir,proxy,cache_dir FROM settings WHERE id=1").Scan(&downloadDir, &proxy, &cacheDir); err != nil {
			s.failTask(id, err)
			return
		}
		if downloadDir == "" {
			downloadDir = s.providerConfig.OutputDir
		}
		if s.provider.Name() == "mock" {
			if cacheDir == "" {
				cacheDir = filepath.Join(s.root, "cache")
			}
			if err := os.MkdirAll(cacheDir, 0700); err != nil {
				s.failTask(id, err)
				return
			}
		}
		if err := os.MkdirAll(downloadDir, 0700); err != nil {
			s.failTask(id, err)
			return
		}
		dst := filepath.Join(downloadDir, req.VideoID+"."+req.Kind+map[string]string{"video": ".mp4", "audio": ".m4a"}[req.Kind])
		provider := s.provider
		if s.provider.Name() == "yt-dlp" {
			provider = YtDlpProvider{Binary: os.Getenv("YT_DLP_BINARY"), OutputDir: downloadDir, Proxy: proxy}
			dst = ""
		}
		result, err := provider.Download(ctx, req, dst, offset, func(p Progress) {
			_, _ = s.db.Exec("UPDATE download_tasks SET downloaded_bytes=?,total_bytes=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", p.Downloaded, p.Total, id)
		})
		if ctx.Err() != nil {
			_, _ = s.db.Exec("UPDATE download_tasks SET status='paused',updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
			s.mu.Lock()
			s.workerStates[id] = "paused"
			s.mu.Unlock()
			return
		}
		if err != nil {
			_, _ = s.db.Exec("UPDATE download_tasks SET status='failed',error=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", err.Error(), id)
			s.mu.Lock()
			s.workerStates[id] = "failed"
			s.mu.Unlock()
			return
		}
		if result.Path == "" {
			_, _ = s.db.Exec("UPDATE download_tasks SET status='failed',error='provider returned no media path' WHERE id=?", id)
			s.mu.Lock()
			s.workerStates[id] = "failed"
			s.mu.Unlock()
			return
		}
		column := "video_path"
		if req.Kind == "audio" {
			column = "audio_path"
		}
		if _, err = s.db.Exec("UPDATE videos SET "+column+"=? WHERE id=?", result.Path, req.VideoID); err != nil {
			_, _ = s.db.Exec("UPDATE download_tasks SET status='failed',error=? WHERE id=?", err.Error(), id)
			s.mu.Lock()
			s.workerStates[id] = "failed"
			s.mu.Unlock()
			return
		}
		_, _ = s.db.Exec("UPDATE download_tasks SET status='done',downloaded_bytes=total_bytes,error=NULL,updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
		s.mu.Lock()
		s.workerStates[id] = "done"
		s.mu.Unlock()
	}()
}
func (s *Server) failTask(id int64, err error) {
	_, _ = s.db.Exec("UPDATE download_tasks SET status='failed',error=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", err.Error(), id)
	s.mu.Lock()
	s.workerStates[id] = "failed"
	s.mu.Unlock()
}
func (s *Server) pause(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		fail(w, 400, "invalid_id", "invalid task id")
		return
	}
	s.mu.Lock()
	cancel := s.cancels[id]
	state := s.workerStates[id]
	s.mu.Unlock()
	if cancel == nil || state != "running" {
		fail(w, 409, "invalid_state", "task is not running")
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
		fail(w, 409, "invalid_state", "task finished before pause completed")
		return
	}
	jsonOut(w, map[string]string{"status": "paused"})
}
func (s *Server) resume(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		fail(w, 400, "invalid_id", "invalid task id")
		return
	}
	var req DownloadRequest
	var status string
	if err = s.db.QueryRow("SELECT video_id,kind,status FROM download_tasks WHERE id=?", id).Scan(&req.VideoID, &req.Kind, &status); err == sql.ErrNoRows {
		fail(w, 404, "not_found", "task not found")
		return
	} else if err != nil {
		fail(w, 500, "database", err.Error())
		return
	}
	if status != "paused" && status != "queued" {
		fail(w, 409, "invalid_state", "only queued or paused tasks can resume")
		return
	}
	s.start(id, req)
	jsonOut(w, map[string]string{"status": "running"})
}
func (s *Server) sync(w http.ResponseWriter, r *http.Request) {
	var favorites, videos int
	if err := s.syncFavorites(r.Context()); err != nil {
		fail(w, 503, "provider_unavailable", err.Error())
		return
	}
	_ = s.db.QueryRow("SELECT (SELECT count(*) FROM favorites),(SELECT count(*) FROM videos)").Scan(&favorites, &videos)
	jsonOut(w, map[string]any{"favorites": favorites, "videos": videos})
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
	kind := r.URL.Query().Get("kind")
	column := "video_path"
	if kind == "audio" {
		column = "audio_path"
	} else if kind != "" && kind != "video" {
		fail(w, 400, "invalid_kind", "kind must be video or audio")
		return
	}
	var path string
	e := s.db.QueryRow("SELECT COALESCE("+column+",'') FROM videos WHERE id=?", id).Scan(&path)
	if e != nil || path == "" {
		fail(w, 404, "media_not_found", "downloaded video is unavailable")
		return
	}
	http.ServeFile(w, r, path)
}
