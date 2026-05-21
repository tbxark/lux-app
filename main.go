package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/tbxark/lux-app/internal/luxrunner"
)

const appID = "github.com.tbxark.lux-app"

var version = "dev"

func main() {
	a := fyneapp.NewWithID(appID)
	w := a.NewWindow("Lux Downloader " + version)
	w.Resize(fyne.NewSize(780, 640))

	prefs := a.Preferences()
	cfg := loadConfig(prefs)

	urls := widget.NewMultiLineEntry()
	urls.SetPlaceHolder("Paste one or more video URLs, one per line.")
	urls.SetMinRowsVisible(4)
	urls.SetText(cfg.URLText)

	outputPath := widget.NewEntry()
	outputPath.SetPlaceHolder(defaultOutputPath())
	outputPath.SetText(cfg.OutputPath)
	browse := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		picker := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri != nil {
				outputPath.SetText(uri.Path())
			}
		}, w)
		picker.Show()
	})
	outputRow := container.NewBorder(nil, nil, nil, browse, outputPath)

	stream := widget.NewEntry()
	stream.SetPlaceHolder("Optional, for example 1080, 720, best stream id")
	stream.SetText(cfg.Stream)

	cookie := widget.NewMultiLineEntry()
	cookie.SetPlaceHolder("Optional Cookie header or cookie file path supported by lux")
	cookie.SetMinRowsVisible(2)
	cookie.SetText(cfg.Cookie)

	userAgent := widget.NewEntry()
	userAgent.SetPlaceHolder("Optional User-Agent")
	userAgent.SetText(cfg.UserAgent)

	refer := widget.NewEntry()
	refer.SetPlaceHolder("Optional Referer")
	refer.SetText(cfg.Refer)

	retry := widget.NewEntry()
	retry.SetText(fmt.Sprint(cfg.Retry))
	threads := widget.NewEntry()
	threads.SetText(fmt.Sprint(cfg.Threads))
	chunkSize := widget.NewEntry()
	chunkSize.SetText(fmt.Sprint(cfg.ChunkSizeMB))

	playlist := widget.NewCheck("Download playlist", nil)
	playlist.SetChecked(cfg.Playlist)
	audioOnly := widget.NewCheck("Audio only", nil)
	audioOnly.SetChecked(cfg.AudioOnly)
	caption := widget.NewCheck("Download captions", nil)
	caption.SetChecked(cfg.Caption)
	multiThread := widget.NewCheck("Multi-thread download", nil)
	multiThread.SetChecked(cfg.MultiThread)
	silent := widget.NewCheck("Silent lux output", nil)
	silent.SetChecked(cfg.Silent)

	logOutput := widget.NewMultiLineEntry()
	logOutput.SetMinRowsVisible(9)
	logOutput.Wrapping = fyne.TextWrapWord
	logOutput.Disable()

	var (
		logMu sync.Mutex
		lines []string
	)
	appendLog := func(format string, args ...any) {
		line := fmt.Sprintf(format, args...)
		stamped := time.Now().Format("15:04:05") + "  " + line
		logMu.Lock()
		lines = append(lines, stamped)
		text := strings.Join(lines, "\n")
		logMu.Unlock()
		fyne.Do(func() {
			logOutput.SetText(text)
		})
	}

	status := widget.NewLabel("Ready")
	progress := widget.NewProgressBar()
	progress.Hide()
	progressDetail := widget.NewLabel("")
	progressDetail.Wrapping = fyne.TextWrapWord
	progressDetail.Hide()

	var (
		progressLogMu   sync.Mutex
		lastProgressLog string
	)
	updateProgress := func(event luxrunner.ProgressEvent) {
		if line := progressLogLine(event); line != "" {
			progressLogMu.Lock()
			if line != lastProgressLog {
				lastProgressLog = line
				appendLog("%s", line)
			}
			progressLogMu.Unlock()
		}
		fyne.Do(func() {
			applyProgressEvent(status, progress, progressDetail, event)
		})
	}

	var runButton *widget.Button
	runButton = widget.NewButtonWithIcon("Download", theme.DownloadIcon(), func() {
		cfg := downloadConfig{
			URLText:     urls.Text,
			URLs:        parseURLs(urls.Text),
			OutputPath:  strings.TrimSpace(outputPath.Text),
			Stream:      stream.Text,
			Cookie:      cookie.Text,
			UserAgent:   userAgent.Text,
			Refer:       refer.Text,
			Playlist:    playlist.Checked,
			AudioOnly:   audioOnly.Checked,
			Caption:     caption.Checked,
			MultiThread: multiThread.Checked,
			Silent:      silent.Checked,
			Retry:       parsePositiveInt(retry.Text, 10),
			Threads:     parsePositiveInt(threads.Text, 10),
			ChunkSizeMB: parsePositiveInt(chunkSize.Text, 1),
		}
		if cfg.OutputPath == "" {
			cfg.OutputPath = defaultOutputPath()
			outputPath.SetText(cfg.OutputPath)
		}
		if len(cfg.URLs) == 0 {
			dialog.ShowInformation("Missing URL", "Paste at least one URL before starting a download.", w)
			return
		}

		saveConfig(prefs, cfg)
		runButton.Disable()
		progressLogMu.Lock()
		lastProgressLog = ""
		progressLogMu.Unlock()
		progress.SetValue(0)
		progress.Show()
		progressDetail.SetText("Waiting for download metadata")
		progressDetail.Show()
		status.SetText("Preparing")
		appendLog("starting %d URL(s)", len(cfg.URLs))

		go func() {
			err := luxrunner.Run(cfg.runnerConfig(updateProgress))
			fyne.Do(func() {
				progress.Hide()
				progressDetail.Hide()
				runButton.Enable()
				if err != nil {
					status.SetText("Failed")
					appendLog("failed: %v", err)
					dialog.ShowError(err, w)
					return
				}
				status.SetText("Finished")
				appendLog("finished")
			})
		}()
	})
	runButton.Importance = widget.HighImportance

	downloadForm := widget.NewForm(
		widget.NewFormItem("URLs", urls),
		widget.NewFormItem("Output folder", outputRow),
		widget.NewFormItem("Stream", stream),
	)
	configForm := widget.NewForm(
		widget.NewFormItem("Cookie", cookie),
		widget.NewFormItem("User-Agent", userAgent),
		widget.NewFormItem("Referer", refer),
		widget.NewFormItem("Retry", retry),
		widget.NewFormItem("Threads", threads),
		widget.NewFormItem("Chunk size MB", chunkSize),
	)
	checks := container.NewGridWithColumns(2, playlist, audioOnly, caption, multiThread, silent)

	content := container.NewBorder(
		nil,
		container.NewVBox(
			container.NewBorder(nil, nil, status, runButton, progress),
			progressDetail,
		),
		nil,
		nil,
		container.NewVScroll(container.NewVBox(
			downloadForm,
			widget.NewSeparator(),
			checks,
			widget.NewSeparator(),
			configForm,
			widget.NewSeparator(),
			container.NewBorder(nil, nil, widget.NewLabel("Log"), layout.NewSpacer(), nil),
			logOutput,
		)),
	)

	w.SetContent(content)
	w.ShowAndRun()
}

func applyProgressEvent(status *widget.Label, progress *widget.ProgressBar, detail *widget.Label, event luxrunner.ProgressEvent) {
	switch event.Phase {
	case luxrunner.ProgressExtracting:
		status.SetText("Extracting")
		progress.SetValue(0)
	case luxrunner.ProgressDownloading:
		if event.Total > 0 {
			progress.SetValue(event.Percent)
			status.SetText(fmt.Sprintf("Downloading %.1f%%", event.Percent*100))
		} else {
			status.SetText("Downloading")
		}
	case luxrunner.ProgressMerging:
		progress.SetValue(1)
		status.SetText("Merging")
	case luxrunner.ProgressSkipped:
		progress.SetValue(1)
		status.SetText("Skipped")
	case luxrunner.ProgressFinished:
		progress.SetValue(1)
		status.SetText("Finished")
	}
	detail.SetText(progressDetailText(event))
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
