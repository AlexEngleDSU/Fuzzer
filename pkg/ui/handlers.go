package ui

import (
	"context"
	"fmt"
	"strconv"
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
	state *engine.ScanState,
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
		fyne.Do(func() { statusLabel.SetText("Scanner initializing...") })
		
		*followMode = true
		
		if *cancelFuncPtr != nil { (*cancelFuncPtr)() }
		
		var ctx context.Context
		ctx, *cancelFuncPtr = context.WithCancel(context.Background())

		state.Mu.Lock()
		state.Results = []engine.ScanResult{}
		state.Mu.Unlock()
		list.Refresh()

		depth, _ := strconv.Atoi(depthEntry.Text)
		wordlist, _ := engine.ReadLines(pathEntry.Text)
		threads, _ := strconv.Atoi(threadEntry.Text)
		delayS, _ := strconv.Atoi(delayEntry.Text)

		go func() {
			fyne.Do(func() { statusLabel.SetText("Launching browser...") })
			session, err := browser.InitializeSession(urlEntry.Text)
			    
			if err != nil {
				// This will print the EXACT error to your terminal
				fmt.Printf("CRITICAL BROWSER ERROR: %v\n", err)
				statusLabel.SetText("Browser Error! Check terminal.")
				return
			} 
                	if session != nil && len(session.Cookies) > 0 {
			        engine.InjectCookies(engine.GlobalJar, urlEntry.Text, session.Cookies)
			        fyne.Do(func() { statusLabel.SetText("WAF Solved - Cookies Injected - Starting Scan!") })
			} else {
				statusLabel.SetText("Warning: No cookies found, scanning anyway...")
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
			    if filterEntry.Text != "" && strconv.Itoa(res.StatusCode) == filterEntry.Text { continue }
			    
			    displayStr := fmt.Sprintf("Status: %d | URL: %s", res.StatusCode, res.URL)
			    if res.Location != "" { displayStr += fmt.Sprintf(" -> Redirect: %s", res.Location)}

			    state.Add(res)
			    fyne.Do(func() { list.Refresh() })
			}
		}()

		fyne.Do(func() {
			tabs.SelectIndex(1)
			tabs.Refresh()
		})
	}
}
