package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
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

type downloadConfig struct {
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

func defaultOutputPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, "Downloads")
}

func loadConfig(prefs fyne.Preferences) downloadConfig {
	cfg := downloadConfig{
		URLText:     prefs.String(prefURLs),
		OutputPath:  prefs.StringWithFallback(prefOutputPath, defaultOutputPath()),
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
	cfg.URLs = parseURLs(cfg.URLText)
	return cfg
}

func saveConfig(prefs fyne.Preferences, cfg downloadConfig) {
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

func parseURLs(raw string) []string {
	fields := strings.Fields(raw)
	urls := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			urls = append(urls, field)
		}
	}
	return urls
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func (cfg downloadConfig) luxArgs() []string {
	args := []string{"lux"}

	appendValue := func(flag, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		args = append(args, flag, value)
	}
	appendBool := func(flag string, enabled bool) {
		if enabled {
			args = append(args, flag)
		}
	}

	appendValue("--output-path", cfg.OutputPath)
	appendValue("--stream-format", cfg.Stream)
	appendValue("--cookie", cfg.Cookie)
	appendValue("--user-agent", cfg.UserAgent)
	appendValue("--refer", cfg.Refer)

	appendBool("--playlist", cfg.Playlist)
	appendBool("--audio-only", cfg.AudioOnly)
	appendBool("--caption", cfg.Caption)
	appendBool("--multi-thread", cfg.MultiThread)
	appendBool("--silent", cfg.Silent)

	if cfg.Retry > 0 {
		args = append(args, "--retry", strconv.Itoa(cfg.Retry))
	}
	if cfg.Threads > 0 {
		args = append(args, "--thread", strconv.Itoa(cfg.Threads))
	}
	if cfg.ChunkSizeMB > 0 {
		args = append(args, "--chunk-size", strconv.Itoa(cfg.ChunkSizeMB))
	}

	args = append(args, cfg.URLs...)
	return args
}

func redactLuxArgs(args []string) []string {
	redacted := append([]string(nil), args...)
	for i := 0; i < len(redacted)-1; i++ {
		if redacted[i] == "--cookie" {
			redacted[i+1] = "[redacted]"
			i++
		}
	}
	return redacted
}
