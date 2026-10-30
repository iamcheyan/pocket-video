package main

import (
	"context"
	"os"
	"path/filepath"
)

// start launches one download worker. A task already owned by a worker is never
// started twice, and the worker always leaves a stable terminal state behind.
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
