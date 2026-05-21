package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	"github.com/tbxark/lux-app/internal/ui/view"
	"github.com/tbxark/lux-app/internal/ui/viewmodel"
)

const appID = "github.com.tbxark.lux-app"

func main() {
	a := fyneapp.NewWithID(appID)
	meta := a.Metadata()
	w := a.NewWindow(fmt.Sprintf("Lux Downloader v%s (%d)", meta.Version, meta.Build))
	w.Resize(fyne.NewSize(780, 860))

	vm := viewmodel.New(a.Preferences())
	w.SetContent(view.Build(w, vm))
	w.ShowAndRun()
}
