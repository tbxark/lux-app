package main

import (
	"reflect"
	"testing"
)

func TestParseURLs(t *testing.T) {
	got := parseURLs(" https://example.com/a\nhttps://example.com/b\tBV1xx ")
	want := []string{"https://example.com/a", "https://example.com/b", "BV1xx"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseURLs() = %#v, want %#v", got, want)
	}
}

func TestLuxArgs(t *testing.T) {
	cfg := downloadConfig{
		URLs:        []string{"https://example.com/video"},
		OutputPath:  "/tmp/videos",
		Stream:      "720",
		Cookie:      "a=b",
		UserAgent:   "lux-app-test",
		Refer:       "https://example.com",
		Playlist:    true,
		AudioOnly:   true,
		Caption:     true,
		MultiThread: true,
		Silent:      true,
		Retry:       3,
		Threads:     4,
		ChunkSizeMB: 8,
	}

	got := cfg.luxArgs()
	want := []string{
		"lux",
		"--output-path", "/tmp/videos",
		"--stream-format", "720",
		"--cookie", "a=b",
		"--user-agent", "lux-app-test",
		"--refer", "https://example.com",
		"--playlist",
		"--audio-only",
		"--caption",
		"--multi-thread",
		"--silent",
		"--retry", "3",
		"--thread", "4",
		"--chunk-size", "8",
		"https://example.com/video",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("luxArgs() = %#v, want %#v", got, want)
	}
}

func TestRedactLuxArgs(t *testing.T) {
	got := redactLuxArgs([]string{"lux", "--cookie", "secret", "--retry", "3", "https://example.com"})
	want := []string{"lux", "--cookie", "[redacted]", "--retry", "3", "https://example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("redactLuxArgs() = %#v, want %#v", got, want)
	}
}
