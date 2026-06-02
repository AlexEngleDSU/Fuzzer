package ui

import (
	"fmt"
	"image/color"
	"net/url"
	"strconv"
	"strings"
	"sync"


	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
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

type fixedEntry struct {
	*widget.Entry
	width float32
}

func (f *fixedEntry) MinSize() fyne.Size {
	return fyne.NewSize(f.width, f.Entry.MinSize().Height)
}

type myTheme struct{ fyne.Theme }

func (m myTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameForeground || n == theme.ColorNameHyperlink {
		return color.White
	}
	return m.Theme.Color(n, v)
}

func StartGUI() {
	fmt.Println("Initializing Application...")
	a := app.NewWithID("com.fuzzer.app")
	fmt.Println("App object created.")
	a.Settings().SetTheme(&myTheme{Theme: theme.DefaultTheme()})
	w := a.NewWindow("Fuzzer GUI")
	fmt.Println("Window object created.")
	w.Resize(fyne.NewSize(700, 500))


	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://example.com/FUZZ")

	recursiveCheck := widget.NewCheck("", nil)

	depthEntry := widget.NewEntry()
	depthEntry.SetText("3")

	filterEntry := widget.NewEntry()
	// Replace the filterContainer definition with this:
	filterContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(150, 40)), filterEntry)
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
			resChan := engine.ConcurrentScan(urlEntry.Text, wordlist, 10, filterEntry.Text, recursiveCheck.Checked, depth)
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


        optionsRow := container.NewHBox(
        	widget.NewLabel("Recursive Check"),
                recursiveCheck,
                widget.NewLabel("Depth:"),
                depthEntry,
                widget.NewLabel("Filter:"),
                filterContainer,
        )

        // Now set the window content
        w.SetContent(container.NewBorder(
                container.NewVBox(
                        urlEntry,
                        optionsRow,
                        container.NewHBox(selectButton, pathLabel),
                        startButton,
                ),
                nil, nil, nil, list,
        ))

        w.ShowAndRun()
}
