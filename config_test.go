package main

import (
	"reflect"
	"testing"
)

func TestParseURLs(t *testing.T) {
	got := parseURLs(" https://example.com/a\nhttps://example.com/b\n  BV1xx  \n\n")
	want := []string{"https://example.com/a", "https://example.com/b", "BV1xx"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseURLs() = %#v, want %#v", got, want)
	}
}

func TestRunnerConfig(t *testing.T) {
	cfg := downloadConfig{
		URLs:        []string{"https://example.com/video"},
		OutputPath:  "/tmp/videos",
		Stream:      " 720 ",
		Cookie:      " a=b ",
		UserAgent:   " lux-app-test ",
		Refer:       " https://example.com ",
		Playlist:    true,
		AudioOnly:   true,
		Caption:     true,
		MultiThread: true,
		Silent:      true,
		Retry:       3,
		Threads:     4,
		ChunkSizeMB: 8,
	}

	got := cfg.runnerConfig(nil)
	if !reflect.DeepEqual(got.URLs, cfg.URLs) {
		t.Fatalf("runnerConfig().URLs = %#v, want %#v", got.URLs, cfg.URLs)
	}
	if got.OutputPath != "/tmp/videos" ||
		got.Stream != "720" ||
		got.Cookie != "a=b" ||
		got.UserAgent != "lux-app-test" ||
		got.Refer != "https://example.com" ||
		!got.Playlist ||
		!got.AudioOnly ||
		!got.Caption ||
		!got.MultiThread ||
		!got.Silent ||
		got.RetryTimes != 3 ||
		got.ThreadNumber != 4 ||
		got.ChunkSizeMB != 8 {
		t.Fatalf("runnerConfig() = %#v", got)
	}
}
