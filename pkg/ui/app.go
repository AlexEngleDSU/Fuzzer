package ui

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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

	// UI Components
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://example.com/FUZZ")

	list := widget.NewList(
		func() int { return len(results) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(results[i])
		},
	)

	startButton := widget.NewButton("Start Scan", func() {
		// Clear previous results
		results = []string{}
		list.Refresh()

		// Run in background so UI stays responsive
		go func() {
			// Hardcoded wordlist for testing; you can add a file picker later!
			wordlist := []string{"admin", "config", "login", "uploads"} 
			
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
	    container.NewVBox(inputRow, startButton),
	    nil, nil, nil,
	    list,
	))

	w.ShowAndRun()
}
