package ui

import (
	"fmt"
	"image/color"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"context"
	"time"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/AlexEngleDSU/Fuzzer/pkg/engine"
)

type SelectableEntry struct {
    widget.Entry
    isFocused bool
}

func (m *SelectableEntry) FocusGained() {
	m.Entry.FocusGained()
	fmt.Println("FocusGained() was triggered!")
	m.isFocused = true
	m.Refresh()

	// Use fyne.Do to ensure thread-safe UI execution
	fyne.Do(func() {
		// This tells the entry to execute the standard "Select All" 
		// shortcut command used by the context menus.
		m.TypedShortcut(&fyne.ShortcutSelectAll{})
	})
}

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

	var cancelFunc context.CancelFunc

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://example.com/FUZZ")

	recursiveCheck := widget.NewCheck("", nil)

	depthEntry := &SelectableEntry{}
	depthEntry.ExtendBaseWidget(depthEntry)
	depthEntry.SetText("3")

	filterEntry := &SelectableEntry{}
	filterEntry.ExtendBaseWidget(filterEntry)
	filterEntry.SetPlaceHolder("Enter filter...")
	filterContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(140, 40)), filterEntry)


	threadEntry := &SelectableEntry{}
	threadEntry.ExtendBaseWidget(threadEntry)
	threadEntry.SetText("10")
	threadContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(50, 40)), threadEntry)

	delayEntry := &SelectableEntry{}
	delayEntry.ExtendBaseWidget(delayEntry)
	delayEntry.SetText("0")
	delayContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 40)), delayEntry)

	headerInput := widget.NewMultiLineEntry()
	headerInput.SetPlaceHolder("User-Agent: MyFuzzer\nAuthorization: Bearer token123")
	headerContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 60)), headerInput)

	var wordlistPath string
	pathEntry := &SelectableEntry{}
	pathEntry.ExtendBaseWidget(pathEntry)
	pathEntry.SetPlaceHolder("No wordlist selected")
	pathContainer := container.New(
	    layout.NewBorderLayout(nil, nil, nil, nil), // Padding is handled by the container
	    pathEntry,
	)


	selectButton := widget.NewButton("Select Wordlist", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			wordlistPath = reader.URI().Path()
			pathEntry.SetText(wordlistPath)
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
		if cancelFunc != nil {
			cancelFunc()
		}

		var ctx context.Context
		ctx, cancelFunc = context.WithCancel(context.Background())

		if urlEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("Please select a target first"), w)
		}
		
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

		threads, err := strconv.Atoi(threadEntry.Text)
		if err != nil {
			threads = 10
		}

		delayS, err := strconv.Atoi(delayEntry.Text)
                if err != nil {
	       		delayS = 0 // Default to 0 if input is invalid
	        }

		customHeaders := make(map[string]string)
		lines := strings.Split(headerInput.Text, "\n")
		for _, line := range lines {
		        parts := strings.SplitN(line, ":", 2)
		        if len(parts) == 2 { customHeaders[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1]) }
		}

		go func() {
			resChan := engine.ConcurrentScan(
				ctx,
				urlEntry.Text,
				wordlist,
				threads,
				filterEntry.Text,
				recursiveCheck.Checked,
				depth,
				time.Duration(delayS) * time.Second,
				customHeaders,
			)
			displayed := make(map[string]bool)

			for res := range resChan {
				displayString := ""
				if res.Message != "" {
					displayString = res.Message
				} else {
					displayString = fmt.Sprintf("Status: %d | URL: %s | Len: %d\n", res.StatusCode, res.URL, res.ContentLength)
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
                widget.NewLabel("Filter Status:"),
                filterContainer,
                widget.NewLabel("Threads:"),
                threadContainer,
                widget.NewLabel("Delay (s): "),
                delayContainer,
        )

        optionsRow2 := container.NewHBox(
        	headerContainer,
        )

        pathRow := container.New(
	    layout.NewBorderLayout(nil, nil, selectButton, nil),
	    selectButton,   // Pin to the left
	    pathContainer,  // The pathContainer takes the remaining space in the center
	)

        // Now set the window content
        header := container.NewVBox(
	    urlEntry,
	    optionsRow,
	    optionsRow2,
	    pathRow, // This replaces the old HBox container
	    startButton,
	)

	w.SetContent(container.NewBorder(
	    header, // Use the variable here!
	    nil, nil, nil, 
	    list,
	))

        w.ShowAndRun()
}
