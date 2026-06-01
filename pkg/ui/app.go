package ui

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/layout"
	"github.com/AlexEngleDSU/Fuzzer/pkg/engine"
)

func StartGUI() {
	a := app.New()
	w := a.NewWindow("Fuzzer GUI")
	w.Resize(fyne.NewSize(600, 400))

	// State management
	var results []string
	var mu sync.Mutex // Prevents UI crashes during concurrent updates
	var selectedWordlist string

	// UI Components
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://example.com/FUZZ")

	wordlistLable := widget.NewLabel("Default: /usr/share/wordlists")

	list := widget.NewList(
		func() int { return len(results) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(results[i])
		},
	)


	fileButton := widget.NewButton("Select Wordlist", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil { return }
			selectedWordlist = reader.URI().Path()
			wordlistLable.SetText("Wordlist: " + selectedWordlist)
		}, w)

		defaultPath := "/usr/share/wordlists"
		listable, err := storage.ListerForURI(storage.NewFileURI(defaultPath))
		if err == nil{ fd.SetLocation(listable) }
		// Optional: filter to only show .txt files
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
		fd.Show()
	})

	startButton := widget.NewButton("Start Scan", func() {
		if selectedWordlist == "" {
			dialog.ShowError(fmt.Errorf("Please select a wordlist first"), w)
			return
		}

		wordlist, err := engine.ReadLines(selectedWordlist)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		// Clear previous results
		results = []string{}
		list.Refresh()

		// Run in background so UI stays responsive
		go func() {
			// Hardcoded wordlist for testing; you can add a file picker later!
			resChan := engine.ConcurrentScan(urlEntry.Text, wordlist, 10, "404")
			for res := range resChan {
				mu.Lock()
				msg := fmt.Sprintf("[%d] %s", res.StatusCode, res.URL)
				results = append(results, msg)
				list.Refresh()
				mu.Unlock()
			}
		}()
	})

	// Replace your existing container logic with this:
	// Using FormLayout makes labels and inputs align perfectly
	inputRow := container.New(layout.NewFormLayout(), 
	    widget.NewLabel("Target URL:"), 
	    urlEntry,
	)

	w.SetContent(container.NewBorder(
	    container.NewVBox(
	    	inputRow,
	    	container.NewHBox(fileButton, wordlistLable),
	    	startButton,
	    ),
	    nil, nil, nil,
	    list,
	))


	w.ShowAndRun()
}
