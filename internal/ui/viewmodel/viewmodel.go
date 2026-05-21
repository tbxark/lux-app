package viewmodel

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"

	"github.com/tbxark/lux-app/internal/luxrunner"
	"github.com/tbxark/lux-app/internal/ui/model"
)

type Download struct {
	prefs fyne.Preferences

	URLText    binding.String
	OutputPath binding.String
	Stream     binding.String
	Cookie     binding.String
	UserAgent  binding.String
	Refer      binding.String
	Retry      binding.String
	Threads    binding.String
	ChunkSize  binding.String

	Playlist    binding.Bool
	AudioOnly   binding.Bool
	Caption     binding.Bool
	MultiThread binding.Bool
	Silent      binding.Bool

	Status          binding.String
	LogText         binding.String
	ProgressValue   binding.Float
	ProgressDetail  binding.String
	ProgressVisible binding.Bool
	Running         binding.Bool

	OnError func(error)
	OnInfo  func(title, message string)

	logMu sync.Mutex
	lines []string

	progressLogMu   sync.Mutex
	lastProgressLog string
}

func New(prefs fyne.Preferences) *Download {
	cfg := model.Load(prefs)
	vm := &Download{
		prefs: prefs,

		URLText:    binding.NewString(),
		OutputPath: binding.NewString(),
		Stream:     binding.NewString(),
		Cookie:     binding.NewString(),
		UserAgent:  binding.NewString(),
		Refer:      binding.NewString(),
		Retry:      binding.NewString(),
		Threads:    binding.NewString(),
		ChunkSize:  binding.NewString(),

		Playlist:    binding.NewBool(),
		AudioOnly:   binding.NewBool(),
		Caption:     binding.NewBool(),
		MultiThread: binding.NewBool(),
		Silent:      binding.NewBool(),

		Status:          binding.NewString(),
		LogText:         binding.NewString(),
		ProgressValue:   binding.NewFloat(),
		ProgressDetail:  binding.NewString(),
		ProgressVisible: binding.NewBool(),
		Running:         binding.NewBool(),
	}

	_ = vm.URLText.Set(cfg.URLText)
	_ = vm.OutputPath.Set(cfg.OutputPath)
	_ = vm.Stream.Set(cfg.Stream)
	_ = vm.Cookie.Set(cfg.Cookie)
	_ = vm.UserAgent.Set(cfg.UserAgent)
	_ = vm.Refer.Set(cfg.Refer)
	_ = vm.Retry.Set(fmt.Sprint(cfg.Retry))
	_ = vm.Threads.Set(fmt.Sprint(cfg.Threads))
	_ = vm.ChunkSize.Set(fmt.Sprint(cfg.ChunkSizeMB))
	_ = vm.Playlist.Set(cfg.Playlist)
	_ = vm.AudioOnly.Set(cfg.AudioOnly)
	_ = vm.Caption.Set(cfg.Caption)
	_ = vm.MultiThread.Set(cfg.MultiThread)
	_ = vm.Silent.Set(cfg.Silent)
	_ = vm.Status.Set("Ready")
	return vm
}

func (vm *Download) StartDownload() {
	cfg := vm.snapshot()
	if cfg.OutputPath == "" {
		cfg.OutputPath = model.DefaultOutputPath()
		_ = vm.OutputPath.Set(cfg.OutputPath)
	}
	if len(cfg.URLs) == 0 {
		if vm.OnInfo != nil {
			vm.OnInfo("Missing URL", "Paste at least one URL before starting a download.")
		}
		return
	}

	model.Save(vm.prefs, cfg)

	vm.progressLogMu.Lock()
	vm.lastProgressLog = ""
	vm.progressLogMu.Unlock()

	_ = vm.Running.Set(true)
	_ = vm.ProgressValue.Set(0)
	_ = vm.ProgressDetail.Set("Waiting for download metadata")
	_ = vm.ProgressVisible.Set(true)
	_ = vm.Status.Set("Preparing")
	vm.appendLog("starting %d URL(s)", len(cfg.URLs))

	go func() {
		err := luxrunner.Run(cfg.RunnerConfig(vm.handleProgress))
		fyne.Do(func() {
			_ = vm.ProgressVisible.Set(false)
			_ = vm.Running.Set(false)
			if err != nil {
				_ = vm.Status.Set("Failed")
				vm.appendLog("failed: %v", err)
				if vm.OnError != nil {
					vm.OnError(err)
				}
				return
			}
			_ = vm.Status.Set("Finished")
			vm.appendLog("finished")
		})
	}()
}

func (vm *Download) snapshot() model.Config {
	getString := func(b binding.String) string {
		v, _ := b.Get()
		return v
	}
	getBool := func(b binding.Bool) bool {
		v, _ := b.Get()
		return v
	}

	urlText := getString(vm.URLText)
	return model.Config{
		URLText:     urlText,
		URLs:        model.ParseURLs(urlText),
		OutputPath:  strings.TrimSpace(getString(vm.OutputPath)),
		Stream:      getString(vm.Stream),
		Cookie:      getString(vm.Cookie),
		UserAgent:   getString(vm.UserAgent),
		Refer:       getString(vm.Refer),
		Playlist:    getBool(vm.Playlist),
		AudioOnly:   getBool(vm.AudioOnly),
		Caption:     getBool(vm.Caption),
		MultiThread: getBool(vm.MultiThread),
		Silent:      getBool(vm.Silent),
		Retry:       model.ParsePositiveInt(getString(vm.Retry), 10),
		Threads:     model.ParsePositiveInt(getString(vm.Threads), 10),
		ChunkSizeMB: model.ParsePositiveInt(getString(vm.ChunkSize), 1),
	}
}

func (vm *Download) handleProgress(event luxrunner.ProgressEvent) {
	if line := progressLogLine(event); line != "" {
		vm.progressLogMu.Lock()
		if line != vm.lastProgressLog {
			vm.lastProgressLog = line
			vm.appendLog("%s", line)
		}
		vm.progressLogMu.Unlock()
	}
	fyne.Do(func() {
		vm.applyProgressEvent(event)
	})
}

func (vm *Download) appendLog(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	stamped := time.Now().Format("15:04:05") + "  " + line
	vm.logMu.Lock()
	vm.lines = append(vm.lines, stamped)
	text := strings.Join(vm.lines, "\n")
	vm.logMu.Unlock()
	fyne.Do(func() {
		_ = vm.LogText.Set(text)
	})
}

func (vm *Download) applyProgressEvent(event luxrunner.ProgressEvent) {
	switch event.Phase {
	case luxrunner.ProgressExtracting:
		_ = vm.Status.Set("Extracting")
		_ = vm.ProgressValue.Set(0)
	case luxrunner.ProgressDownloading:
		if event.Total > 0 {
			_ = vm.ProgressValue.Set(event.Percent)
			_ = vm.Status.Set(fmt.Sprintf("Downloading %.1f%%", event.Percent*100))
		} else {
			_ = vm.Status.Set("Downloading")
		}
	case luxrunner.ProgressMerging:
		_ = vm.ProgressValue.Set(1)
		_ = vm.Status.Set("Merging")
	case luxrunner.ProgressSkipped:
		_ = vm.ProgressValue.Set(1)
		_ = vm.Status.Set("Skipped")
	case luxrunner.ProgressFinished:
		_ = vm.ProgressValue.Set(1)
		_ = vm.Status.Set("Finished")
	}
	_ = vm.ProgressDetail.Set(progressDetailText(event))
}

func progressDetailText(event luxrunner.ProgressEvent) string {
	target := event.Title
	if target == "" {
		target = event.FileName
	}
	if target == "" {
		target = event.URL
	}
	if target == "" {
		target = string(event.Phase)
	}

	switch {
	case event.Total > 0:
		return fmt.Sprintf("%s  %s / %s", target, formatBytes(event.Current), formatBytes(event.Total))
	case event.Current > 0:
		return fmt.Sprintf("%s  %s", target, formatBytes(event.Current))
	default:
		return target
	}
}

func progressLogLine(event luxrunner.ProgressEvent) string {
	target := event.Title
	if target == "" {
		target = event.URL
	}
	switch event.Phase {
	case luxrunner.ProgressExtracting:
		return fmt.Sprintf("extracting %s", target)
	case luxrunner.ProgressMerging:
		return fmt.Sprintf("merging %s", target)
	case luxrunner.ProgressSkipped:
		return fmt.Sprintf("skipped %s", target)
	default:
		return ""
	}
}

func formatBytes(value int64) string {
	if value < 1024 {
		return fmt.Sprintf("%d B", value)
	}
	units := []string{"KiB", "MiB", "GiB", "TiB"}
	size := float64(value)
	for _, unit := range units {
		size /= 1024
		if size < 1024 {
			return fmt.Sprintf("%.1f %s", size, unit)
		}
	}
	return fmt.Sprintf("%.1f PiB", size/1024)
}
