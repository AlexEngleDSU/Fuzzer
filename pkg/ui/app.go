package ui

import (
	"fmt"
	"image/color"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/AlexEngleDSU/Fuzzer/pkg/engine"
)

// compactLink forces list rows to be dense (20px height)
type compactLink struct {
	*widget.Hyperlink
}

func (c *compactLink) MinSize() fyne.Size {
	return fyne.NewSize(c.Hyperlink.MinSize().Width, 25)
}

type myTheme struct{ fyne.Theme }

func (m myTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameForeground || n == theme.ColorNameHyperlink {
		return color.White
	}
	return m.Theme.Color(n, v)
}

func StartGUI() {
	a := app.NewWithID("com.fuzzer.app")
	a.Settings().SetTheme(&myTheme{Theme: theme.DefaultTheme()})
	w := a.NewWindow("Fuzzer GUI")
	w.Resize(fyne.NewSize(700, 500))

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://example.com/FUZZ")
	recursiveCheck := widget.NewCheck("Enable Recursion", nil)
	depthEntry := widget.NewEntry()
	depthEntry.SetText("3")

	var wordlistPath string
	pathLabel := widget.NewLabel("No wordlist selected")

	selectButton := widget.NewButton("Select Wordlist", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			wordlistPath = reader.URI().Path()
			pathLabel.SetText("Selected: " + wordlistPath)
		}, w)
	})

	var mu sync.Mutex
	resultsList := []string{}

	list := widget.NewList(
		func() int { return len(resultsList) },
		func() fyne.CanvasObject { return &compactLink{widget.NewHyperlink("", nil)} },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			mu.Lock()
			text := resultsList[id]
			mu.Unlock()

			link := obj.(*compactLink)
			link.SetText(text)

			parts := strings.Fields(text)
			for _, p := range parts {
				if strings.HasPrefix(p, "http") {
					parsed, _ := url.Parse(strings.TrimSuffix(p, "/"))
					link.URL = parsed
					break
				}
			}
		},
	)

	startButton := widget.NewButton("Start Scan", func() {
		if wordlistPath == "" {
			dialog.ShowError(fmt.Errorf("please select a wordlist first"), w)
			return
		}

		mu.Lock()
		resultsList = []string{}
		mu.Unlock()
		list.Refresh()

		depth, _ := strconv.Atoi(depthEntry.Text)
		wordlist, _ := engine.ReadLines(wordlistPath)

		go func() {
			resChan := engine.ConcurrentScan(urlEntry.Text, wordlist, 10, "404", recursiveCheck.Checked, depth)
			displayed := make(map[string]bool)

			for res := range resChan {
				displayString := ""
				if res.Message != "" {
					displayString = res.Message
				} else {
					displayString = fmt.Sprintf("\t[%d] %s", res.StatusCode, res.URL)
					if res.Location != "" {
						displayString += " -> " + res.Location
					}
				}

				mu.Lock()
				if !displayed[res.URL] || res.Message != "" {
					resultsList = append(resultsList, displayString)
					displayed[res.URL] = true
				}
				mu.Unlock()

				fyne.Do(func() {
					list.Refresh()
					list.ScrollToBottom()
				})
			}
		}()
	})

	w.SetContent(container.NewBorder(
		container.NewVBox(urlEntry, container.NewHBox(widget.NewLabel("Options: "), recursiveCheck, depthEntry), container.NewHBox(selectButton, pathLabel), startButton),
		nil, nil, nil, list,
	))
	w.ShowAndRun()
}
