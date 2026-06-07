package engine

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	"strconv"
	fhttp "github.com/bogdanfinn/fhttp"
)

func ConcurrentScan(ctx context.Context, host, urlTemplate, headerTemplate string, wordlist []string, workerCount int, filterCodes string, recursive bool, maxDepth int, delay time.Duration) <-chan ScanResult {
	results := make(chan ScanResult, 500)
	browserClient := CreateBrowserClient()

	// Initial setup
	performHandshake(browserClient, host)
	badContentLength := get404Length(browserClient, urlTemplate)
	headers := GetOrderedHeaders(headerTemplate)
	
	filterMap := make(map[int]bool)
	for _, codeStr := range strings.Split(filterCodes, ",") {
		if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil {
			filterMap[code] = true
		}
	}

	go func() {
		defer close(results)
		queue := []string{urlTemplate}
		lastRedirectURL := &sync.Map{}
		lastRedirectURL.Store("latest", getBaseURL(urlTemplate))

		for depth := 0; depth <= maxDepth; depth++ {
			if len(queue) == 0 { break }

			results <- ScanResult{Message: fmt.Sprintf("[+] Depth %d | Paths: %d", depth, len(queue))}

			nextGen := []string{}
			var mu sync.Mutex
			var globalSeen sync.Map
			var wg sync.WaitGroup
			jobs := make(chan string, 100)

			for w := 1; w <= workerCount; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for target := range jobs {
						// Resume/Pause check
						PauseMu.Lock()
						if Paused {
							pChan := PauseChan
							PauseMu.Unlock()
							<-pChan
						} else {
							PauseMu.Unlock()
						}

						select {
						case <-ctx.Done(): return
						default:
						}

						if _, loaded := globalSeen.LoadOrStore(target, true); loaded { continue }

						req, _ := fhttp.NewRequest("GET", target, nil)
						// Reconstruct headers
						req.Header = fhttp.Header{}
						for _, h := range headers {
							if strings.EqualFold(h.Key, "Host") { req.Host = h.Value; continue }
							req.Header.Add(h.Key, h.Value)
						}
						
						resp, err := browserClient.Do(req)
						if err != nil || resp == nil { continue }
						
						body, _ := io.ReadAll(resp.Body)
						resp.Body.Close()

						status := resp.StatusCode
						if !filterMap[status] && !(status == 200 && int64(len(body)) == badContentLength) {
							results <- ScanResult{URL: target, StatusCode: status, ContentLength: int64(len(body))}

							if recursive && (status >= 300 && status < 400) {
								loc := resp.Header.Get("Location")
								if loc != "" {
									resolved := ResolveURL(target, loc)
									mu.Lock()
									nextGen = append(nextGen, resolved)
									mu.Unlock()
								}
							}
						}
					}
				}()
			}

			for _, base := range queue {
				for _, word := range wordlist {
					jobs <- strings.ReplaceAll(base, "FUZZ", word)
				}
			}
			close(jobs)
			wg.Wait()
			queue = nextGen
		}
	}()
	return results
}
