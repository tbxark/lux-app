package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/tbxark/lux-app/internal/ui/model"
	"github.com/tbxark/lux-app/internal/ui/viewmodel"
)

func Build(window fyne.Window, vm *viewmodel.Download) fyne.CanvasObject {
	urls := widget.NewMultiLineEntry()
	urls.SetPlaceHolder("Paste one or more video URLs, one per line.")
	urls.SetMinRowsVisible(4)
	urls.Bind(vm.URLText)

	outputPath := widget.NewEntryWithData(vm.OutputPath)
	outputPath.SetPlaceHolder(model.DefaultOutputPath())
	browse := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		picker := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if uri != nil {
				_ = vm.OutputPath.Set(uri.Path())
			}
		}, window)
		picker.Show()
	})
	outputRow := container.NewBorder(nil, nil, nil, browse, outputPath)

	stream := widget.NewEntryWithData(vm.Stream)
	stream.SetPlaceHolder("Optional, for example 1080, 720, best stream id")

	cookie := widget.NewMultiLineEntry()
	cookie.SetPlaceHolder("Optional Cookie header or cookie file path supported by lux")
	cookie.SetMinRowsVisible(2)
	cookie.Bind(vm.Cookie)

	userAgent := widget.NewEntryWithData(vm.UserAgent)
	userAgent.SetPlaceHolder("Optional User-Agent")

	refer := widget.NewEntryWithData(vm.Refer)
	refer.SetPlaceHolder("Optional Referer")

	retry := widget.NewEntryWithData(vm.Retry)
	threads := widget.NewEntryWithData(vm.Threads)
	chunkSize := widget.NewEntryWithData(vm.ChunkSize)

	playlist := widget.NewCheckWithData("Download playlist", vm.Playlist)
	audioOnly := widget.NewCheckWithData("Audio only", vm.AudioOnly)
	caption := widget.NewCheckWithData("Download captions", vm.Caption)
	multiThread := widget.NewCheckWithData("Multi-thread download", vm.MultiThread)
	silent := widget.NewCheckWithData("Silent lux output", vm.Silent)

	logOutput := widget.NewMultiLineEntry()
	logOutput.SetMinRowsVisible(9)
	logOutput.Wrapping = fyne.TextWrapWord
	logOutput.Bind(vm.LogText)
	logOutput.Disable()

	status := widget.NewLabelWithData(vm.Status)
	progress := widget.NewProgressBarWithData(vm.ProgressValue)
	progress.Hide()
	progressDetail := widget.NewLabelWithData(vm.ProgressDetail)
	progressDetail.Wrapping = fyne.TextWrapWord
	progressDetail.Hide()

	vm.ProgressVisible.AddListener(binding.NewDataListener(func() {
		visible, _ := vm.ProgressVisible.Get()
		if visible {
			progress.Show()
			progressDetail.Show()
		} else {
			progress.Hide()
			progressDetail.Hide()
		}
	}))

	runButton := widget.NewButtonWithIcon("Download", theme.DownloadIcon(), vm.StartDownload)
	runButton.Importance = widget.HighImportance
	vm.Running.AddListener(binding.NewDataListener(func() {
		running, _ := vm.Running.Get()
		if running {
			runButton.Disable()
		} else {
			runButton.Enable()
		}
	}))

	vm.OnError = func(err error) { dialog.ShowError(err, window) }
	vm.OnInfo = func(title, message string) { dialog.ShowInformation(title, message, window) }

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

	return container.NewBorder(
		nil,
		container.NewVBox(
			container.NewBorder(nil, nil, nil, runButton, status),
			progress,
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
}
