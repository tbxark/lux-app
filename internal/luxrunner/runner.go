package luxrunner

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/iawia002/lux/app"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"

	"github.com/tbxark/lux-app/internal/luxdownloader"
)

type ProgressEvent = luxdownloader.ProgressEvent
type ProgressFunc = luxdownloader.ProgressFunc
type ProgressPhase = luxdownloader.ProgressPhase

const (
	ProgressExtracting  = luxdownloader.ProgressExtracting
	ProgressDownloading = luxdownloader.ProgressDownloading
	ProgressMerging     = luxdownloader.ProgressMerging
	ProgressSkipped     = luxdownloader.ProgressSkipped
	ProgressFinished    = luxdownloader.ProgressFinished
)

type Config struct {
	URLs         []string
	OutputPath   string
	Stream       string
	Cookie       string
	UserAgent    string
	Refer        string
	Playlist     bool
	AudioOnly    bool
	Caption      bool
	MultiThread  bool
	Silent       bool
	RetryTimes   int
	ThreadNumber int
	ChunkSizeMB  int

	OnProgress       ProgressFunc
	ProgressThrottle time.Duration
}

func Run(config Config) error {
	if len(config.URLs) == 0 {
		return errors.New("too few arguments")
	}
	normalize(&config)

	cookie, err := cookieValue(config.Cookie)
	if err != nil {
		return err
	}
	request.SetOptions(request.Options{
		RetryTimes: config.RetryTimes,
		Cookie:     cookie,
		UserAgent:  config.UserAgent,
		Refer:      config.Refer,
		Silent:     config.Silent,
	})

	var errs []error
	for _, rawURL := range config.URLs {
		videoURL := strings.TrimSpace(rawURL)
		if videoURL == "" {
			continue
		}
		emit(config.OnProgress, ProgressEvent{
			Phase:   ProgressExtracting,
			URL:     videoURL,
			Message: "extracting",
		})

		data, err := extractors.Extract(videoURL, extractors.Options{
			Playlist:     config.Playlist,
			ThreadNumber: config.ThreadNumber,
			Cookie:       cookie,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("extract %s: %w", videoURL, err))
			continue
		}

		downloader := luxdownloader.New(luxdownloader.Options{
			Silent:           config.Silent,
			Stream:           config.Stream,
			AudioOnly:        config.AudioOnly,
			Refer:            config.Refer,
			OutputPath:       config.OutputPath,
			FileNameLength:   255,
			Caption:          config.Caption,
			MultiThread:      config.MultiThread,
			ThreadNumber:     config.ThreadNumber,
			RetryTimes:       config.RetryTimes,
			ChunkSizeMB:      config.ChunkSizeMB,
			OnProgress:       config.OnProgress,
			ProgressThrottle: config.ProgressThrottle,
		})
		for _, item := range data {
			if item.Err != nil {
				errs = append(errs, fmt.Errorf("extract %s: %w", item.URL, item.Err))
				continue
			}
			if err := downloader.Download(item); err != nil {
				errs = append(errs, fmt.Errorf("download %s: %w", item.URL, err))
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func normalize(config *Config) {
	if config.RetryTimes <= 0 {
		config.RetryTimes = 10
	}
	if config.ThreadNumber <= 0 {
		config.ThreadNumber = 10
	}
	if config.ChunkSizeMB <= 0 {
		config.ChunkSizeMB = 1
	}
}

func cookieValue(cookie string) (string, error) {
	cookie = strings.TrimSpace(cookie)
	if cookie == "" {
		return "", nil
	}
	info, err := os.Stat(cookie)
	if err != nil || info.IsDir() {
		return cookie, nil
	}
	data, err := os.ReadFile(cookie)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func emit(callback ProgressFunc, event ProgressEvent) {
	if callback != nil {
		callback(event)
	}
}
