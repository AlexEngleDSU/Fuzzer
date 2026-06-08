package ui

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/container"
	"github.com/AlexEngleDSU/Fuzzer/pkg/engine"
	"github.com/AlexEngleDSU/Fuzzer/pkg/browser"
)

type AppState struct {
	Window fyne.Window
	Tabs   container.AppTabs // Requires the widget import above
}

var followMode = new(bool) 

func HandleStartScan(
	tabs *container.AppTabs, 
	list *widget.List,
	resultsList *[]string,
	resultsMu *sync.Mutex,
	urlEntry *widget.Entry,
	userHeaderInput *widget.Entry,
	depthEntry *SelectableEntry,
	threadEntry *SelectableEntry,
	delayEntry *SelectableEntry,
	pathEntry *SelectableEntry,
	filterEntry *SelectableEntry,
	recursiveCheck *widget.Check,
	cancelFuncPtr *context.CancelFunc,
	statusLabel *widget.Label,
	resumeBtn *widget.Button,
	followModePtr *bool,
) func() {
	return func() {
		
		*followMode = true
		
		if *cancelFuncPtr != nil {
			(*cancelFuncPtr)()
		}
		var ctx context.Context
		ctx, *cancelFuncPtr = context.WithCancel(context.Background())

		resultsMu.Lock()
		*resultsList = []string{}
		resultsMu.Unlock()
		list.Refresh()

		depth, _ := strconv.Atoi(depthEntry.Text)
		wordlist, _ := engine.ReadLines(pathEntry.Text)
		threads, _ := strconv.Atoi(threadEntry.Text)
		delayS, _ := strconv.Atoi(delayEntry.Text)

		go func() {
			statusLabel.SetText("Launching browser...")
			session, err := browser.InitializeSession(urlEntry.Text)
			    
			if err != nil {
				// This will print the EXACT error to your terminal
				fmt.Printf("CRITICAL BROWSER ERROR: %v\n", err)
				statusLabel.SetText("Browser Error! Check terminal.")
				return
			} 
                	if session != nil && len(session.Cookies) > 0 {
			        engine.InjectCookies(engine.GlobalJar, urlEntry.Text, session.Cookies)
			        statusLabel.SetText("WAF Solved!")
			} else {
			        statusLabel.SetText("Challenge passed (no cookies).")
    			}
			resChan := engine.ConcurrentScan(
				ctx, 
				extractHost(urlEntry.Text), 
				urlEntry.Text, 
				userHeaderInput.Text, 
				wordlist, threads, 
				filterEntry.Text, 
				recursiveCheck.Checked, 
				depth, 
				time.Duration(delayS)*time.Second,
			)
			for res := range resChan {
			    if filterEntry.Text != "" && strconv.Itoa(res.StatusCode) == filterEntry.Text {
				continue
			    }
			    
			    resultsMu.Lock()
			    *resultsList = append(*resultsList, fmt.Sprintf("Status: %d | URL: %s", res.StatusCode, res.URL))
			    resultsMu.Unlock()

			    // Inside your resChan loop
			    fyne.Do(func() {
 				// Tells Fyne to update the list count/content
				list.Refresh()
				if *followModePtr {
				    list.ScrollToBottom()
				} else {
				    resumeBtn.Show()
				}
			    })
			}
		}()

		fyne.Do(func() {
			tabs.SelectIndex(1)
			tabs.Refresh()
		})
	}
}
