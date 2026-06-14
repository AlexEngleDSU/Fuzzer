package engine

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
)

var GlobalPauser = NewPauser()

func shouldDisplay(status int, filterCodes, matchCodes []int) bool {
    // 1. If Match Codes exist, the status MUST be in the match list to proceed
    if len(matchCodes) > 0 {
        found := false
        for _, code := range matchCodes {
            if code == status {
                found = true
                break
            }
        }
        if !found { return false }
    }

    // 2. If Filter Codes exist, the status MUST NOT be in the filter list
    for _, code := range filterCodes {
        if code == status {
            return false
        }
    }

    return true
}

func ConcurrentScan(
	ctx context.Context, 
	host, 
	urlTemplate, 
	headerTemplate string,
	wordlist []string, 
	workerCount int, 
	matchCodes string, 
	filterCodes string, 
	recursive bool, 
	maxDepth int, 
	delay time.Duration, 
	onStatusUpdate func(string)) <-chan ScanResult {
	
	var wg sync.WaitGroup
	results := make(chan ScanResult, 500)
	jobs := make(chan Job, 2000)
	discovery := make(chan Job, 1000)
	filterCodesMap := make(map[int]bool)
	matchMap := make(map[int]bool)
	hasMatchCodes := len(strings.TrimSpace(matchCodes)) > 0
	if hasMatchCodes {
		for _, codeStr := range strings.Split(matchCodes, ",") {
			if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil { matchMap[code] = true }
		}
	}
	
	for _, codeStr := range strings.Split(filterCodes, ",") {
		if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil {
			filterCodesMap[code] = true
		}
	}

	browserClient := CreateBrowserClient()
	baseHeaders := fmt.Sprintf(headerTemplate, host, host)
	headers := GetOrderedHeaders(baseHeaders)
	badContentLength := get404Length(browserClient, urlTemplate)

	// Single global tracking for seen URLs
	globalSeen := sync.Map{}

	// WORKER POOL
	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				// Single source of truth for deduplication
				if _, loaded := globalSeen.LoadOrStore(j.URL, true); loaded { continue }
				// Apply delay if configured
				if delay > 0 { time.Sleep(delay) }

				req, _ := fhttp.NewRequest("GET", j.URL, nil)
				for _, h := range headers {
					if strings.EqualFold(h.Key, "Host") {
						req.Host = h.Value
						continue
					}
					req.Header.Add(h.Key, h.Value)
				}

				resp, err := browserClient.Do(req)
				if err != nil { continue }

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				status := resp.StatusCode
				isSoft404 := (status == 200 && int64(len(body)) == badContentLength)
				
				if recursive && status >= 300 && status < 400 && j.Depth < maxDepth {
				    resolved := ResolveURL(j.URL, resp.Header.Get("Location"))
				    discovery <- Job{URL: resolved, Depth: j.Depth + 1}
				}
				isFiltered := filterCodesMap[status]
				isNotMatched := hasMatchCodes && !matchMap[status]
				if !isFiltered && !isNotMatched && !isSoft404 {
				    results <- ScanResult{
					URL:           j.URL,
					StatusCode:    status,
					ContentLength: int64(len(body)),
					Location:      resp.Header.Get("Location"),
					Depth:         j.Depth,
				    }
				}
			}
		}()
	}
	// ORCHESTRATOR
	go func() {
		queue := []Job{{URL: urlTemplate, Depth: 0}}
		onStatusUpdate("Starting Scan!")

		for depth := 0; depth <= maxDepth; depth++ {
			for _, base := range queue {
				for _, word := range wordlist {
					target := strings.ReplaceAll(base.URL, "FUZZ", word)
					select {
					    case jobs <- Job{URL: target, Depth: base.Depth}:
						// Successfully queued
					    case <-time.After(5 * time.Second):
						fmt.Println("CRITICAL: Jobs channel full, check worker health!")
						close(jobs)
						wg.Wait()
						close(results)
						return 
					}
				}
			}

			// Wait for current burst to process
			time.Sleep(1 * time.Second)

			// Collect recursive discoveries
			nextQueue := []Job{}
			for i := 0; i < len(discovery); i++ {
				d := <-discovery
				if !strings.Contains(d.URL, "FUZZ") {
					if !strings.HasSuffix(d.URL, "/") { d.URL += "/" }
					d.URL += "FUZZ"
				}
				nextQueue = append(nextQueue, d)
			}
			queue = nextQueue
			if len(queue) == 0 { break }
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	return results
}
