package engine

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ScanResult struct {
	URL        string
	StatusCode int
	Location   string
	Message    string
}

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil { return nil, err }
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() { lines = append(lines, scanner.Text()) }
	return lines, scanner.Err()
}

func resolveURL(base, loc string) string {
	if strings.HasPrefix(loc, "http") { return loc }
	if strings.HasPrefix(loc, "//") { return "https:" + loc }
	parsedBase, _ := url.Parse(base)
	cleanLoc := strings.TrimLeft(loc, "/")
	return fmt.Sprintf("%s://%s/%s", parsedBase.Scheme, parsedBase.Host, cleanLoc)
}

func ConcurrentScan(ctx context.Context, urlTemplate string, wordlist []string, workerCount int, filterCodes string, recursive bool, maxDepth int, delay time.Duration) <-chan ScanResult {
	results := make(chan ScanResult, 500)
	filterMap := make(map[int]bool)
	for _, codeStr := range strings.Split(filterCodes, ",") {
		if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil {
			filterMap[code] = true
		}
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	go func() {
		// Queue starts with the initial user-provided template
		queue := []string{urlTemplate}

		for depth := 0; depth <= maxDepth; depth++ {
			if len(queue) == 0 { break }

			results <- ScanResult{Message: fmt.Sprintf("[+] Starting depth level: %d | Base Paths: %d", depth, len(queue))}

			// Findings for next depth
			nextGen := []string{}
			var mu sync.Mutex
			var globalSeen sync.Map
			var wg sync.WaitGroup

			// Job channel for the wordlist expansion
			jobs := make(chan string, 100)

			// Spawn workers
			for w := 1; w <= workerCount; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					var ticker *time.Ticker
					if delay > 0 {
						ticker := time.NewTicker(delay)
						defer ticker.Stop()
					}
					for target := range jobs {
						select {
						case <-ctx.Done():
							return
						default:
							if ticker != nil {
								<-ticker.C
							}
						}
						if _, loaded := globalSeen.LoadOrStore(target, true); loaded { continue }

						resp, err := client.Get(target)
						if err != nil { continue }

						status := resp.StatusCode
						loc := resp.Header.Get("Location")
						resp.Body.Close()

						if !filterMap[status] {
							results <- ScanResult{URL: target, StatusCode: status, Location: loc}

							if recursive && (status >= 300 && status < 400) && loc != "" {
								resolved := resolveURL(target, loc)
								newPath := strings.TrimSuffix(resolved, "/") + "/FUZZ"
								mu.Lock()
								nextGen = append(nextGen, newPath)
								mu.Unlock()
							}
						}
					}
				}()
			}

			// Feed workers: for every base in the queue, run the whole wordlist
			for _, base := range queue {
				for _, word := range wordlist {
					finalURL := strings.ReplaceAll(base, "FUZZ", word)
					jobs <- finalURL
				}
			}
			close(jobs)
			wg.Wait()

			// Reset queue to discovered paths
			queue = nextGen
		}
		close(results)
	}()
	return results
}
