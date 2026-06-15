package appUI
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
	"github.com/AlexEngleDSU/Fuzzer/pkg/browser"
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
	browser.EnsureEnvironment()
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
	    myUI,
            myUI.URLEntry, 
            myUI.PathEntry,
            myUI.RecursiveCheck,
            myUI.DepthEntry, 
            myUI.ThreadEntry, 
            myUI.DelayEntry,
            myUI.TimeoutEntry, 
            myUI.MatchCodesEntry,
            myUI.FilterCodesEntry,
            myUI.UserHeaderInput,
        )
	// 4. Layout
	ScanControlsGrid := container.NewHBox(
		container.New(layout.NewGridLayout(6),
		    widget.NewLabel("Recursion:"), 
		    container.NewCenter(myUI.RecursiveCheck), 
		    widget.NewLabel("Depth:"), 
		    myUI.DepthEntry,
		    layout.NewSpacer(),
		    layout.NewSpacer(),
		    
		    widget.NewLabel("Threads:"), 
		    myUI.ThreadEntry, 
		    widget.NewLabel("Delay:"), 
		    myUI.DelayEntry,
		    widget.NewLabel("Timeout:"),
		    myUI.TimeoutEntry,
		    
		    widget.NewLabel("Match Codes:"), 
		    myUI.MatchCodesEntry, 
		    widget.NewLabel("Filter Codes:"), 
		    myUI.FilterCodesEntry,
		),
	)
		
	ScanControls := container.NewVBox(
	    myUI.URLEntry,
	    container.NewBorder(nil, nil, selectButton, nil, myUI.PathEntry),
	    ScanControlsGrid,
	    myUI.StartButton,
	    widget.NewLabel("Initial Request"),
	)
	// Tab Content
	
	ConfigControls := container.NewVBox(
	    widget.NewLabel("Scan Mode:"),
	    myUI.StatefulRadio,
	    widget.NewSeparator(),
	    container.NewHBox(widget.NewLabel("Clear CookieJar:"), myUI.ClearJarCheck),
	    widget.NewLabel("Cookie Allow List (comma-separated):"),
	     // This will now take up the full width of the VBox
	)
	
	ConfigContent := container.NewBorder(ConfigControls, container.NewVBox(myUI.ScanButton,), nil, nil,  myUI.CookieAllowList)
	
	myUI.ScanButton.OnTapped = func() {
	    ctrl.Tabs.SelectIndex(1)
	    ctrl.Tabs.Refresh()
	}
	
	
	ScanContent := container.NewBorder(ScanControls, nil, nil, nil, myUI.UserHeaderInput)
	
	ctrl.Tabs = container.NewAppTabs(
		container.NewTabItem("Config", ConfigContent),
		container.NewTabItem("Scan", ScanContent),
		container.NewTabItem("Results", SetupResultsContent(&ctrl)),
		container.NewTabItem("Output", container.NewHBox(widget.NewLabel("NOT IMPLEMENTED YET!!!"))),
	)
	
	width, height := screen.GetPrimaryScreenSize()
//	w.Resize(fyne.NewSize(width-10, height-75))
	width = 700
	height = 800
	w.Resize(fyne.NewSize(width, height))
	w.CenterOnScreen()
	w.SetContent(ctrl.Tabs)
	w.ShowAndRun()
}
