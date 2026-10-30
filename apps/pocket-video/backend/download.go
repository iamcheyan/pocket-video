package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ProviderDownloadConfig struct{ OutputDir, Proxy string }

func (MockProvider) Download(ctx context.Context, req DownloadRequest, dst string, offset int64, progress func(Progress)) (Result, error) {
	source := dst + ".source" + filepath.Ext(dst)
	if _, err := os.Stat(source); err != nil {
		if offset != 0 {
			return Result{}, fmt.Errorf("resume source missing for task at byte %d", offset)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
			return Result{}, err
		}
		if err := generateDemoMedia(ctx, source, req.Kind); err != nil {
			return Result{}, err
		}
	}
	in, err := os.Open(source)
	if err != nil {
		return Result{}, err
	}
	defer in.Close()
	st, err := in.Stat()
	if err != nil {
		return Result{}, err
	}
	total := st.Size()
	if offset < 0 || offset > total {
		return Result{}, fmt.Errorf("invalid resume offset %d for %d-byte media", offset, total)
	}
	if _, err = in.Seek(offset, io.SeekStart); err != nil {
		return Result{}, err
	}
	flags := os.O_CREATE | os.O_WRONLY
	if offset == 0 {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}
	out, err := os.OpenFile(dst, flags, 0600)
	if err != nil {
		return Result{}, err
	}
	defer out.Close()
	if offset == 0 {
		progress(Progress{0, total})
	}
	buf := make([]byte, 64*1024)
	downloaded := offset
	for {
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		default:
		}
		n, re := in.Read(buf)
		if n > 0 {
			wn, we := out.Write(buf[:n])
			if we != nil {
				return Result{}, we
			}
			if wn != n {
				return Result{}, io.ErrShortWrite
			}
			downloaded += int64(wn)
			if err := out.Sync(); err != nil {
				return Result{}, err
			}
			progress(Progress{downloaded, total})
			time.Sleep(120 * time.Millisecond)
		}
		if re == io.EOF {
			break
		}
		if re != nil {
			return Result{}, re
		}
	}
	if err = out.Sync(); err != nil {
		return Result{}, err
	}
	_ = os.Remove(source)
	return Result{Path: dst}, nil
}

func generateDemoMedia(ctx context.Context, dst, kind string) error {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("mock playable media requires ffmpeg: %w", err)
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y"}
	if kind == "audio" {
		args = append(args, "-f", "lavfi", "-i", "sine=frequency=440:duration=8", "-c:a", "aac", "-movflags", "+faststart", dst)
	} else {
		args = append(args, "-f", "lavfi", "-i", "testsrc2=size=640x360:rate=24:duration=8", "-f", "lavfi", "-i", "sine=frequency=440:duration=8", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-movflags", "+faststart", dst)
	}
	output, err := exec.CommandContext(ctx, ffmpeg, args...).CombinedOutput()
	if err != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("generate mock media: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

type YtDlpProvider struct{ Binary, OutputDir, Proxy string }

func (y YtDlpProvider) Name() string { return "yt-dlp" }
func (y YtDlpProvider) SyncFavorites(context.Context) ([]Favorite, error) {
	return nil, fmt.Errorf("Bilibili login is not configured; yt-dlp cannot sync favorites")
}
func (y YtDlpProvider) ListVideos(context.Context, Query) ([]Video, error) {
	return nil, fmt.Errorf("Bilibili login is not configured; yt-dlp cannot list favorites")
}
func (y YtDlpProvider) Download(ctx context.Context, req DownloadRequest, dst string, offset int64, progress func(Progress)) (Result, error) {
	binary := y.Binary
	if binary == "" {
		binary = "yt-dlp"
	}
	if err := os.MkdirAll(y.OutputDir, 0700); err != nil {
		return Result{}, err
	}
	template := filepath.Join(y.OutputDir, "%(id)s.%(ext)s")
	args := []string{"--newline", "--progress", "--continue", "-o", template}
	if y.Proxy != "" {
		args = append(args, "--proxy", y.Proxy)
	}
	if req.Kind == "audio" {
		args = append(args, "-f", "bestaudio[ext=m4a]/bestaudio", "--extract-audio", "--audio-format", "m4a")
	} else {
		args = append(args, "-f", "best[ext=mp4]/bestvideo+bestaudio/best", "--merge-output-format", "mp4")
	}
	args = append(args, "https://www.bilibili.com/video/"+req.VideoID)
	cmd := exec.CommandContext(ctx, binary, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, err
	}
	cmd.Stderr = cmd.Stdout
	if err = cmd.Start(); err != nil {
		return Result{}, fmt.Errorf("start yt-dlp: %w", err)
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[download]") {
			got, total, ok := parseYtProgress(line)
			if ok {
				progress(Progress{offset + got, offset + total})
			}
		}
	}
	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if scanErr != nil {
		return Result{}, scanErr
	}
	if waitErr != nil {
		return Result{}, fmt.Errorf("yt-dlp: %w", waitErr)
	}
	matches, _ := filepath.Glob(filepath.Join(y.OutputDir, req.VideoID+".*"))
	for _, path := range matches {
		if strings.HasSuffix(path, ".part") {
			continue
		}
		st, e := os.Stat(path)
		if e == nil && st.Size() > 0 {
			return Result{Path: path}, nil
		}
	}
	return Result{}, fmt.Errorf("yt-dlp completed without an output file for %s", req.VideoID)
}
func parseYtProgress(line string) (int64, int64, bool) {
	// Typical line: [download]  12.5% of 8.00MiB at 1.00MiB/s ETA 00:07
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return 0, 0, false
	}
	percent := strings.TrimSuffix(fields[1], "%")
	pct, err := strconv.ParseFloat(percent, 64)
	if err != nil {
		return 0, 0, false
	}
	var sizeToken string
	for i, f := range fields {
		if f == "of" && i+1 < len(fields) {
			sizeToken = strings.Trim(fields[i+1], "~")
			break
		}
	}
	total, err := parseBytes(sizeToken)
	if err != nil || total == 0 {
		return 0, 0, false
	}
	return int64(float64(total) * pct / 100), total, true
}
func parseBytes(token string) (int64, error) {
	token = strings.Trim(token, "~")
	units := map[string]float64{"B": 1, "KiB": 1024, "MiB": 1024 * 1024, "GiB": 1024 * 1024 * 1024, "KB": 1000, "MB": 1e6, "GB": 1e9}
	for unit, mul := range units {
		if strings.HasSuffix(token, unit) {
			n, e := strconv.ParseFloat(strings.TrimSuffix(token, unit), 64)
			return int64(n * mul), e
		}
	}
	return 0, fmt.Errorf("unknown byte unit %q", token)
}
