package model

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"

	"github.com/tbxark/lux-app/internal/luxrunner"
)

const (
	prefURLs        = "urls"
	prefOutputPath  = "output_path"
	prefStream      = "stream"
	prefCookie      = "cookie"
	prefUserAgent   = "user_agent"
	prefRefer       = "refer"
	prefPlaylist    = "playlist"
	prefAudioOnly   = "audio_only"
	prefCaption     = "caption"
	prefMultiThread = "multi_thread"
	prefSilent      = "silent"
	prefRetry       = "retry"
	prefThreads     = "threads"
	prefChunkSize   = "chunk_size"
)

type Config struct {
	URLText     string
	URLs        []string
	OutputPath  string
	Stream      string
	Cookie      string
	UserAgent   string
	Refer       string
	Playlist    bool
	AudioOnly   bool
	Caption     bool
	MultiThread bool
	Silent      bool
	Retry       int
	Threads     int
	ChunkSizeMB int
}

func DefaultOutputPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, "Downloads")
}

func Load(prefs fyne.Preferences) Config {
	cfg := Config{
		URLText:     prefs.String(prefURLs),
		OutputPath:  prefs.StringWithFallback(prefOutputPath, DefaultOutputPath()),
		Stream:      prefs.String(prefStream),
		Cookie:      prefs.String(prefCookie),
		UserAgent:   prefs.String(prefUserAgent),
		Refer:       prefs.String(prefRefer),
		Playlist:    prefs.Bool(prefPlaylist),
		AudioOnly:   prefs.Bool(prefAudioOnly),
		Caption:     prefs.Bool(prefCaption),
		MultiThread: prefs.BoolWithFallback(prefMultiThread, true),
		Silent:      prefs.Bool(prefSilent),
		Retry:       prefs.IntWithFallback(prefRetry, 10),
		Threads:     prefs.IntWithFallback(prefThreads, 10),
		ChunkSizeMB: prefs.IntWithFallback(prefChunkSize, 1),
	}
	cfg.URLs = ParseURLs(cfg.URLText)
	return cfg
}

func Save(prefs fyne.Preferences, cfg Config) {
	prefs.SetString(prefURLs, cfg.URLText)
	prefs.SetString(prefOutputPath, cfg.OutputPath)
	prefs.SetString(prefStream, cfg.Stream)
	prefs.SetString(prefCookie, cfg.Cookie)
	prefs.SetString(prefUserAgent, cfg.UserAgent)
	prefs.SetString(prefRefer, cfg.Refer)
	prefs.SetBool(prefPlaylist, cfg.Playlist)
	prefs.SetBool(prefAudioOnly, cfg.AudioOnly)
	prefs.SetBool(prefCaption, cfg.Caption)
	prefs.SetBool(prefMultiThread, cfg.MultiThread)
	prefs.SetBool(prefSilent, cfg.Silent)
	prefs.SetInt(prefRetry, cfg.Retry)
	prefs.SetInt(prefThreads, cfg.Threads)
	prefs.SetInt(prefChunkSize, cfg.ChunkSizeMB)
}

func ParseURLs(raw string) []string {
	lines := strings.Split(raw, "\n")
	urls := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			urls = append(urls, line)
		}
	}
	return urls
}

func ParsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func (cfg Config) RunnerConfig(onProgress luxrunner.ProgressFunc) luxrunner.Config {
	return luxrunner.Config{
		URLs:             append([]string(nil), cfg.URLs...),
		OutputPath:       cfg.OutputPath,
		Stream:           strings.TrimSpace(cfg.Stream),
		Cookie:           strings.TrimSpace(cfg.Cookie),
		UserAgent:        strings.TrimSpace(cfg.UserAgent),
		Refer:            strings.TrimSpace(cfg.Refer),
		Playlist:         cfg.Playlist,
		AudioOnly:        cfg.AudioOnly,
		Caption:          cfg.Caption,
		MultiThread:      cfg.MultiThread,
		Silent:           cfg.Silent,
		RetryTimes:       cfg.Retry,
		ThreadNumber:     cfg.Threads,
		ChunkSizeMB:      cfg.ChunkSizeMB,
		OnProgress:       onProgress,
		ProgressThrottle: 100 * time.Millisecond,
	}
}
