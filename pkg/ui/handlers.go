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
)

type AppState struct {
	Window fyne.Window
	Tabs   container.AppTabs // Requires the widget import above
}

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
) func() {
	return func() {
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
			resChan := engine.ConcurrentScan(ctx, extractHost(urlEntry.Text), urlEntry.Text, userHeaderInput.Text, wordlist, threads, filterEntry.Text, recursiveCheck.Checked, depth, time.Duration(delayS)*time.Second)
			for res := range resChan {
				if filterEntry.Text != "" && strconv.Itoa(res.StatusCode) == filterEntry.Text {
					continue
				}
				resultsMu.Lock()
				*resultsList = append(*resultsList, fmt.Sprintf("Status: %d | URL: %s", res.StatusCode, res.URL))
				resultsMu.Unlock()
				fyne.Do(func() {
					list.Refresh()
					list.ScrollToBottom()
				})
			}
		}()

		fyne.Do(func() {
			tabs.SelectIndex(1)
			tabs.Refresh()
		})
	}
}
