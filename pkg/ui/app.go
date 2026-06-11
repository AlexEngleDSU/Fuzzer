package ui

import (
	"net/url"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/AlexEngleDSU/Fuzzer/pkg/engine"
	"github.com/AlexEngleDSU/Fuzzer/pkg/screen"
	"github.com/sqweek/dialog"
)

func StartGUI() {
	ctrl := AppController{
		State:      &engine.ScanState{},
		FollowMode: new(bool),
	}
	
	ctrl.ResultsList = widget.NewList(
            func() int { return 0 }, // Placeholder for now
            func() fyne.CanvasObject { return widget.NewLabel("Template") },
            func(id widget.ListItemID, obj fyne.CanvasObject) {},
        )   

	// Environment setup
	if !engine.BrowserExists() {
		engine.EnsureEnvironment()
	}
	*ctrl.FollowMode = true

	a := app.NewWithID("com.fuzzer.app")
	a.Settings().SetTheme(&myTheme{Theme: theme.DefaultTheme()})
	w := a.NewWindow("Fuzzer GUI")

	// 1. Initialize the UI components from the struct
	myUI := NewFuzzerUI()

	// 2. Logic for URL changes
	myUI.URLEntry.OnChanged = func(newURL string) {
		baseReferer := strings.TrimSuffix(newURL, "/FUZZ")
		if !strings.HasSuffix(baseReferer, "/") { baseReferer += "/" }
		
		u, err := url.Parse(newURL)
		if err != nil { return }
	
		myUI.UpdateHeader(u.Host, baseReferer)
	}

	// 3. Setup File Selection
	selectButton := widget.NewButton("Select Wordlist", func() {
		d := dialog.File().Title("Select Wordlist")
		lastFile := engine.GetLastFilePath()
		if lastDir := filepath.Dir(lastFile); lastDir != "" { d.SetStartDir(lastDir) }
		if filename, err := d.Load(); err == nil {
			myUI.PathEntry.SetText(filename)
			engine.SaveLastFilePath(filename)
		}
	})
	
	myUI.StartButton.OnTapped = ctrl.HandleStartScan(
            myUI.URLEntry, 
            myUI.PathEntry,
            myUI.RecursiveCheck,
            myUI.DepthEntry, 
            myUI.ThreadEntry, 
            myUI.DelayEntry, 
            myUI.FilterCodesEntry,
            myUI.UserHeaderInput,
        )
	

	// 4. Layout
	topControls := container.NewVBox(
		myUI.URLEntry,
		container.NewBorder(nil, nil, selectButton, nil, myUI.PathEntry),
		container.NewHBox(
			widget.NewLabel("Recursion:"), myUI.RecursiveCheck,
			widget.NewLabel("Depth:"), myUI.DepthEntry,
			widget.NewLabel("Filter codes:"), container.New(layout.NewGridWrapLayout(fyne.NewSize(150,40)), myUI.FilterCodesEntry),
			widget.NewLabel("Match Codes:"), container.New(layout.NewGridWrapLayout(fyne.NewSize(150,40)), myUI.MatchCodesEntry),
			widget.NewLabel("Threads:"), container.New(layout.NewGridWrapLayout(fyne.NewSize(50, 40)), myUI.ThreadEntry),
			widget.NewLabel("Delay:"), container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 40)), myUI.DelayEntry),
		),
		myUI.StartButton,
		widget.NewLabel("Initial Request"),
	)

	// Tab Content
	configContent := container.NewBorder(topControls, nil, nil, nil, myUI.UserHeaderInput)
	
	ctrl.Tabs = container.NewAppTabs(
		container.NewTabItem("Configuration", configContent),
		container.NewTabItem("Results", SetupResultsContent(&ctrl)),
		container.NewTabItem("Output", container.NewHBox(widget.NewLabel("NOT IMPLEMENTED YET!!!"))),
	)
	width, height := screen.GetPrimaryScreenSize()
	w.Resize(fyne.NewSize(width-10, height-75))
	w.CenterOnScreen()
	w.SetContent(ctrl.Tabs)
	w.ShowAndRun()
}
