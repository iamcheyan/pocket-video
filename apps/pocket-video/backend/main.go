package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Server owns the database, the provider adapter, and the running download
// workers. Every HTTP handler is a thin translation layer over these.
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

// seed installs the deterministic development catalog on first start. A live
// provider is never seeded with demo rows.
func (s *Server) seed() {
	var n int
	_ = s.db.QueryRow("SELECT count(*) FROM favorites").Scan(&n)
	if n == 0 && s.provider.Name() == "mock" {
		ctx := context.Background()
		favorites, _ := (MockProvider{}).SyncFavorites(ctx)
		for _, x := range favorites {
			_, _ = s.db.Exec("INSERT OR IGNORE INTO favorites(id,title,video_count,last_sync_at) VALUES(?,?,?,?)", x.ID, x.Title, x.VideoCount, x.LastSyncAt)
		}
		videos, _ := (MockProvider{}).ListVideos(ctx, Query{})
		for _, v := range videos {
			_, _ = s.db.Exec("INSERT OR IGNORE INTO videos(id,favorite_id,title,creator,duration,cover) VALUES(?,?,?,?,?,?)", v.ID, v.FavoriteID, v.Title, v.Creator, v.Duration, v.Cover)
		}
	}
	s.refreshFavoriteCounts()
	_, _ = s.db.Exec("INSERT OR IGNORE INTO settings(id,language,theme,cache_dir,download_dir,sync_minutes,auto_sync,proxy) VALUES(1,'en','system',?,?,60,0,'')", filepath.Join(s.root, "cache"), filepath.Join(s.root, "downloads"))
}

func (s *Server) routes(m *http.ServeMux) {
	m.HandleFunc("GET /api/health", s.health)
	m.HandleFunc("GET /api/account", s.account)
	m.HandleFunc("GET /api/favorites", s.favorites)
	m.HandleFunc("GET /api/videos", s.videos)
	m.HandleFunc("GET /api/videos/{id}", s.video)
	m.HandleFunc("GET /api/task", s.tasks)
	m.HandleFunc("GET /api/task/{id}", s.task)
	m.HandleFunc("POST /api/download", s.download)
	m.HandleFunc("POST /api/task/{id}/pause", s.pause)
	m.HandleFunc("POST /api/task/{id}/resume", s.resume)
	m.HandleFunc("DELETE /api/task/{id}", s.deleteTask)
	m.HandleFunc("POST /api/sync", s.sync)
	m.HandleFunc("GET /api/settings", s.settings)
	m.HandleFunc("PUT /api/settings", s.updateSettings)
	m.HandleFunc("GET /api/media/{id}", s.media)
	m.HandleFunc("GET /api/cover/{id}", s.cover)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Range")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonOut(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, code, msg string) {
	w.WriteHeader(status)
	jsonOut(w, map[string]any{"error": map[string]string{"code": code, "message": msg}})
}
