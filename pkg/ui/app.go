package ui

import (
	"context"
	"fmt"
	"image/color"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/AlexEngleDSU/Fuzzer/pkg/engine"
)

// --- Custom Widgets ---
type SelectableEntry struct {
	widget.Entry
	isFocused bool
}

func (m *SelectableEntry) FocusGained() {
	m.Entry.FocusGained()
	m.isFocused = true
	fyne.Do(func() { m.Refresh() })
	fyne.Do(func() { m.TypedShortcut(&fyne.ShortcutSelectAll{}) })
}

type compactLink struct{ *widget.Hyperlink }

func (c *compactLink) MinSize() fyne.Size { return fyne.NewSize(c.Hyperlink.MinSize().Width, 25) }

type myTheme struct{ fyne.Theme }

func (m myTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameForeground || n == theme.ColorNameHyperlink {
		return color.White
	}
	return m.Theme.Color(n, v)
}

func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Host
}

func StartGUI() {
	a := app.NewWithID("com.fuzzer.app")
	a.Settings().SetTheme(&myTheme{Theme: theme.DefaultTheme()})
	w := a.NewWindow("Fuzzer GUI")
	w.Resize(fyne.NewSize(700, 500))

	var cancelFunc context.CancelFunc
	var wordlistPath string

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://example.com/FUZZ")

	recursiveCheck := widget.NewCheck("", nil)
	depthEntry := &SelectableEntry{}
	depthEntry.ExtendBaseWidget(depthEntry)
	depthEntry.SetText("3")

	filterEntry := &SelectableEntry{}
	filterEntry.ExtendBaseWidget(filterEntry)
	filterContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(140, 40)), filterEntry)
	
	threadEntry := &SelectableEntry{}
	threadEntry.ExtendBaseWidget(threadEntry)
	threadEntry.SetText("10")
	threadContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(50, 40)), threadEntry)
	
	delayEntry := &SelectableEntry{}
	delayEntry.ExtendBaseWidget(delayEntry)
	delayEntry.SetText("1")
	delayContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 40)), delayEntry)

	userHeaderInput := widget.NewMultiLineEntry()
	userHeaderInput.ExtendBaseWidget(userHeaderInput)
	headerTemplate := `Host: %s
Sec-Ch-Ua: "Not-A.Brand";v="24", "Chromium";v="146"
Sec-Ch-Ua-Mobile: ?0
Sec-Ch-Ua-Platform: "Linux"
Upgrade-Insecure-Requests: 1
User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7
Sec-Fetch-Site: none
Sec-Fetch-Mode: navigate
Sec-Fetch-User: ?1
Sec-Fetch-Dest: document
Accept-Encoding: gzip, deflate, br
Accept-Language: en-US,en;q=0.9
Referer: %s
Priority: u=0, i
Connection: keep-alive`
	headerBox := container.NewMax(userHeaderInput)

	urlEntry.OnChanged = func(newURL string) {
	    // 1. Prepare Referer: Trim /FUZZ and ensure it ends with a trailing slash
	    baseReferer := strings.TrimSuffix(newURL, "/FUZZ")
	    if !strings.HasSuffix(baseReferer, "/") {
		baseReferer += "/"
	    }

	    // 2. Parse the URL
	    u, err := url.Parse(newURL)
	    if err != nil {
		return
	    }

	    // 3. Prepare GET Path: Take the full path, remove the /FUZZ suffix
	    // If the path was /FUZZ, this makes it /
	    path := u.Path
	    path = strings.TrimSuffix(path, "/FUZZ")
	    if path == "" {
		path = "/"
	    }

	    // 4. Fill the template: [GET Path, Host, Referer]
	    userHeaderInput.SetText(fmt.Sprintf(headerTemplate, u.Host, baseReferer))
	}
		
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
			if idx := strings.Index(text, "http"); idx != -1 {
				urlPart := text[idx:]
				if spaceIdx := strings.Index(urlPart, " "); spaceIdx != -1 { urlPart = urlPart[:spaceIdx] }
				if parsed, err := url.Parse(strings.TrimSuffix(urlPart, "/")); err == nil { link.URL = parsed }
			}
		},
	)

	var isPaused bool
	pauseButton := widget.NewButton("Pause", nil)
	pauseButton.OnTapped = func() {
		isPaused = !isPaused
		engine.SetPause(isPaused)
		if isPaused { pauseButton.SetText("Resume") } else { pauseButton.SetText("Pause") }
	}

	pathEntry := &SelectableEntry{}
	pathEntry.ExtendBaseWidget(pathEntry)
	selectButton := widget.NewButton("Select Wordlist", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				wordlistPath = reader.URI().Path()
				pathEntry.SetText(wordlistPath)
			}
		}, w)
	})
	pathRow := container.NewBorder(nil, nil, selectButton, nil, pathEntry)

	resultsContent := container.NewBorder(pauseButton, nil, nil, nil, list)
	resultsTab := container.NewTabItem("Results", resultsContent)
	tabs := container.NewAppTabs(container.NewTabItem("Configuration", container.NewVBox()), resultsTab)
	
	startButton := widget.NewButton("Start Scan", func() {
		if cancelFunc != nil { cancelFunc() }
		var ctx context.Context
		ctx, cancelFunc = context.WithCancel(context.Background())
		
		// Parse Headers
		customHeaders := make(map[string]string)
		for _, line := range strings.Split(userHeaderInput.Text, "\n") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 { customHeaders[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1]) }
		}

		mu.Lock()
		resultsList = []string{}
		mu.Unlock()
		fyne.Do(func() {list.Refresh()})

		depth, _ := strconv.Atoi(depthEntry.Text)
		wordlist, _ := engine.ReadLines(wordlistPath)
		threads, _ := strconv.Atoi(threadEntry.Text)
		delayS, _ := strconv.Atoi(delayEntry.Text)

		go func() {
			// Filtering logic passed to engine
			resChan := engine.ConcurrentScan(
			    ctx, 
			    extractHost(urlEntry.Text), 
			    urlEntry.Text, 
			    userHeaderInput.Text, // Pass the raw string template here
			    wordlist, 
			    threads, 
			    filterEntry.Text, 
			    recursiveCheck.Checked, 
			    depth, 
			    time.Duration(delayS)*time.Second,
			)

			for res := range resChan {
				// Internal Filter Check
				if filterEntry.Text != "" && strconv.Itoa(res.StatusCode) == filterEntry.Text {
					continue
				}
				mu.Lock()
				resultsList = append(resultsList, fmt.Sprintf("Status: %d | URL: %s", res.StatusCode, res.URL))
				mu.Unlock()
				fyne.Do(func() { list.Refresh(); list.ScrollToBottom() })
			}
			resultsTab.Text = "Results •"
			fyne.Do(func() { tabs.Refresh() })
		}()
		tabs.SelectIndex(1)
	})

	topControls := container.NewVBox(
	    urlEntry,
	    container.NewHBox(
		widget.NewLabel("Rec:"), recursiveCheck, 
		widget.NewLabel("Depth:"), depthEntry, 
		widget.NewLabel("Filter Code:"), filterContainer, // Note: ensure filterEntry is used directly
		widget.NewLabel("Threads:"), threadContainer,
		widget.NewLabel("Delay (s):"), delayContainer,
	    ),
	    pathRow,
	    startButton,
	    widget.NewLabel("Initial Request:"),
	)

	// 2. Use a Border layout to pin the controls to the TOP 
	// and let the headers box expand to fill the rest of the tab area
	configContent := container.NewBorder(
	    topControls, // Top
	    nil,         // Bottom
	    nil,         // Left
	    nil,         // Right
	    headerBox,   // Center (This will stretch to fill all remaining height)
	)

	tabs.Items[0].Content = configContent

	w.SetContent(tabs)
	w.ShowAndRun()
}
