package appUI
import (
	"context"
	"fmt"
	"strconv"
	"time"
	"net/url"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/AlexEngleDSU/Fuzzer/pkg/engine"
	"github.com/AlexEngleDSU/Fuzzer/pkg/browser"
	"github.com/bogdanfinn/tls-client"
)

func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil { return rawURL }
	return u.Host
}

func (ctrl *AppController) HandleStartScan(
    ui *FuzzerUI,
    urlEntry *widget.Entry,
    pathEntry *SelectableEntry,
    recursiveCheck *widget.Check,
    depthEntry *SelectableEntry,
    threadEntry *SelectableEntry,
    delayEntry *SelectableEntry,
    timeoutEntry *SelectableEntry,
    matchCodesEntry *SelectableEntry,
    filterCodesEntry *SelectableEntry,
    userHeaderInput *widget.Entry,
)func() {
	return func() {
		ctrl.State.Mu.Lock()
		ctrl.State.Results = []engine.ScanResult{}
		ctrl.State.Mu.Unlock()
		
		fyne.Do(func() { ctrl.StatusLabel.SetText("Scanner initializing...") })
		
		*ctrl.FollowMode = true
		
		if ctrl.CancelFunc != nil { ctrl.CancelFunc() }
		
		ctx, cancel := context.WithCancel(context.Background())
		ctrl.CancelFunc = cancel
		depth, _ := strconv.Atoi(depthEntry.Text)
		wordlist, _ := engine.ReadLines(pathEntry.Text)
		threads, _ := strconv.Atoi(threadEntry.Text)
		delayS, _ := strconv.Atoi(delayEntry.Text)
		
		timeoutInput := timeoutEntry.Text
		if timeoutInput == "" { timeoutInput = "5" }
		timeoutSec, _ := strconv.Atoi(timeoutEntry.Text)
		timeout := time.Duration(timeoutSec) * time.Second
		
		updateStatus := func(msg string) {
		    fyne.Do(func() {
		        if ctrl.StatusLabel != nil { ctrl.StatusLabel.SetText(msg) }
		    })
		}
		
		options := ui.GetOptions()
		
		go func() {
			if options.ClearJarOnStart { engine.GlobalJar = tls_client.NewCookieJar() }
			
			if options.Stateful {
				
				fyne.Do(func() { ctrl.StatusLabel.SetText("Launching browser...") })
				session, err := browser.InitializeSession(urlEntry.Text)
				if err != nil {
					// This will print the EXACT error to your terminal
					fmt.Printf("CRITICAL BROWSER ERROR: %v\n", err)
					fyne.Do(func() { ctrl.StatusLabel.SetText("Browser Error! Check terminal.") })
					return
				} 
				if session != nil && len(session.Cookies) > 0 {
				    // Convert []map[string]interface{} to []browser.Cookie
				    convertedCookies := make([]browser.Cookie, 0, len(session.Cookies))
				    for _, c := range session.Cookies {
					name, _ := c["name"].(string)
					value, _ := c["value"].(string)
					if name != "" {
					    convertedCookies = append(convertedCookies, browser.Cookie{
						Name:  name,
						Value: value,
					    })
					}
				    }
				    browser.InjectCookies(engine.GlobalJar, urlEntry.Text, convertedCookies)
				    allCookiesMap := engine.GlobalJar.GetAllCookies()
				    for _ , cookies := range allCookiesMap {
					for _, c := range cookies {
						cookieStatus := fmt.Sprintf("%s: %s\n", c.Name, c.Value)
						fyne.Do(func() { ctrl.StatusLabel.SetText(cookieStatus) })
					}
				    }
				    //fyne.Do(func() { ctrl.StatusLabel.SetText("WAF Solved - Cookies Injected") })
				} 
			} else { 
				engine.GlobalJar = tls_client.NewCookieJar()
				fyne.Do(func() { ctrl.StatusLabel.SetText("Warning: No cookies found, scanning anyway...") }) 
			}
			
			resChan := engine.ConcurrentScan(
				options,
				ctx, 
				extractHost(urlEntry.Text), 
				urlEntry.Text,
				userHeaderInput.Text, 
				wordlist, 
				recursiveCheck.Checked, 
				depth, 
				threads,
				time.Duration(delayS)*time.Second, 
				time.Duration(timeout)*time.Second,
				matchCodesEntry.Text,
				filterCodesEntry.Text, 
				updateStatus,
			)
			
			fyne.Do(func() { ctrl.StatusLabel.SetText("Starting Scan!") })
			time.Sleep(1 * time.Second)
			fyne.Do(func() { ctrl.StatusLabel.Hide() })
			
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			
			for res := range resChan {
			    ctrl.State.Add(res)
			    select {
			    case <-ticker.C:
				    fyne.Do(func() { 
				    	ctrl.ResultsList.Refresh() 
				    	if *ctrl.FollowMode { ctrl.ResultsList.ScrollToBottom() }	
		     	    	    })
		     	    default:
		     	    }
			}
			fyne.Do(func() { ctrl.ResultsList.Refresh() })
		}()
		fyne.Do(func() {
			ctrl.Tabs.SelectIndex(2)
			ctrl.Tabs.Refresh()
		})
	}
}
