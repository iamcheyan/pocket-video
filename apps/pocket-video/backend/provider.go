package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// demoFavorites is the deterministic development catalog. It is intentionally
// small, mixed-language content so search, filters, and i18n can be exercised
// without a Bilibili account.
var demoFavorites = []struct {
	ID    string
	Title string
}{
	{"fav-pjs", "PocketJS Deep Dives"},
	{"fav-audio", "Offline Audio Sessions"},
	{"fav-ui", "Handheld UI Studies"},
	{"fav-music", "Bilibili 收藏 · 音乐"},
}

var demoVideos = []Video{
	{"BV1DEMO001", "fav-pjs", "PocketVideo Architecture Demo", "PocketVideo Lab", 305, "", false, false, "", ""},
	{"BV1DEMO002", "fav-pjs", "Offline First: A Practical Guide", "PocketVideo Lab", 428, "", false, false, "", ""},
	{"BV1DEMO003", "fav-pjs", "PocketJS UI on Small Screens", "PocketVideo Lab", 247, "", false, false, "", ""},
	{"BV1DEMO004", "fav-pjs", "REST Contracts Between UI and Service", "Hikari Studio", 512, "", false, false, "", ""},
	{"BV1DEMO005", "fav-audio", "深夜のピアノ · 1 Hour Session", "Night Piano", 3600, "", false, false, "", ""},
	{"BV1DEMO006", "fav-audio", "雨音と環境音 · Ambient Rain", "Sound Atelier", 2700, "", false, false, "", ""},
	{"BV1DEMO007", "fav-audio", "Lo-fi Study Loop No. 3", "Café Tape", 1800, "", false, false, "", ""},
	{"BV1DEMO008", "fav-ui", "Designing for 320×240 Handhelds", "Tiny Screen Lab", 396, "", false, false, "", ""},
	{"BV1DEMO009", "fav-ui", "Directional Focus and Gamepad Navigation", "Tiny Screen Lab", 342, "", false, false, "", ""},
	{"BV1DEMO010", "fav-ui", "Dark Mode Without Pure Black", "Tiny Screen Lab", 284, "", false, false, "", ""},
	{"BV1DEMO011", "fav-music", "夜に駆ける · Piano Cover", "Rui", 268, "", false, false, "", ""},
	{"BV1DEMO012", "fav-music", "城市夜行 · Live Session", "城市夜行乐队", 431, "", false, false, "", ""},
}

// MockProvider serves the deterministic development catalog and writes
// playable demo media with ffmpeg. It never touches the network.
type MockProvider struct{}

func (MockProvider) Name() string { return "mock" }

func (MockProvider) Account(context.Context) (Account, error) {
	return Account{
		Mode:          "demo",
		Provider:      "mock",
		Authenticated: true,
		DisplayName:   "Demo Account",
		Collections:   len(demoFavorites),
	}, nil
}

func (MockProvider) SyncFavorites(context.Context) ([]Favorite, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	out := make([]Favorite, 0, len(demoFavorites))
	for _, f := range demoFavorites {
		count := 0
		for _, v := range demoVideos {
			if v.FavoriteID == f.ID {
				count++
			}
		}
		out = append(out, Favorite{ID: f.ID, Title: f.Title, VideoCount: count, LastSyncAt: now})
	}
	return out, nil
}

func (MockProvider) ListVideos(_ context.Context, q Query) ([]Video, error) {
	needle := strings.ToLower(strings.TrimSpace(q.Q))
	out := make([]Video, 0, len(demoVideos))
	for _, v := range demoVideos {
		if q.FavoriteID != "" && v.FavoriteID != q.FavoriteID {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(v.Title+" "+v.Creator), needle) {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

// YtDlpProvider shells out to yt-dlp. Metadata sync needs a Bilibili login and
// therefore reports an actionable error until credentials are configured.
type YtDlpProvider struct{ Binary, OutputDir, Proxy string }

func (y YtDlpProvider) Name() string { return "yt-dlp" }

func (y YtDlpProvider) Account(context.Context) (Account, error) {
	return Account{Mode: "live", Provider: "yt-dlp", Authenticated: false, DisplayName: "Bilibili"}, nil
}

func (y YtDlpProvider) SyncFavorites(context.Context) ([]Favorite, error) {
	return nil, fmt.Errorf("Bilibili login is not configured; yt-dlp cannot sync favorites")
}

func (y YtDlpProvider) ListVideos(context.Context, Query) ([]Video, error) {
	return nil, fmt.Errorf("Bilibili login is not configured; yt-dlp cannot list favorites")
}
