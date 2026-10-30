package main

import "context"

// Favorite is one Bilibili favorite folder known to the service.
type Favorite struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	VideoCount int    `json:"video_count"`
	LastSyncAt string `json:"last_sync_at,omitempty"`
}

// Video is the metadata the service keeps for one favorite entry. Media paths
// are only present after a task finished successfully.
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

// Task is one download job. Title is joined from videos so clients can render a
// queue without a second lookup.
type Task struct {
	ID              int64  `json:"id"`
	VideoID         string `json:"video_id"`
	Title           string `json:"title,omitempty"`
	Kind            string `json:"kind"`
	Status          string `json:"status"`
	DownloadedBytes int64  `json:"downloaded_bytes"`
	TotalBytes      int64  `json:"total_bytes"`
	Error           string `json:"error,omitempty"`
	Progress        int    `json:"progress"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// Settings mirrors the single settings row and is the full API payload.
type Settings struct {
	Language    string `json:"language"`
	Theme       string `json:"theme"`
	CacheDir    string `json:"cache_dir"`
	DownloadDir string `json:"download_dir"`
	SyncMinutes int    `json:"sync_minutes"`
	AutoSync    bool   `json:"auto_sync"`
	Proxy       string `json:"proxy"`
}

// Account describes the active adapter. The frontend never talks to Bilibili
// directly; it renders whatever mode the service reports.
type Account struct {
	Mode          string `json:"mode"`
	Provider      string `json:"provider"`
	Authenticated bool   `json:"authenticated"`
	DisplayName   string `json:"display_name"`
	Collections   int    `json:"collections"`
}

type Query struct {
	Q          string
	FavoriteID string
}

type DownloadRequest struct {
	VideoID string `json:"video_id"`
	Kind    string `json:"kind"`
}

type Progress struct{ Downloaded, Total int64 }

type Result struct{ Path string }

// VideoProvider is the adapter boundary. MockProvider and YtDlpProvider are the
// current implementations; a BBDownProvider can be added without touching HTTP
// handlers, SQLite code, or the UI.
type VideoProvider interface {
	Name() string
	Account(context.Context) (Account, error)
	SyncFavorites(context.Context) ([]Favorite, error)
	ListVideos(context.Context, Query) ([]Video, error)
	Download(context.Context, DownloadRequest, string, int64, func(Progress)) (Result, error)
}
