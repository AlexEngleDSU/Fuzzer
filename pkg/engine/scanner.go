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
	"io"
)

type ScanResult struct {
	URL        string
	StatusCode int
	Location   string
	Message    string
	ContentLength int64
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

func get404Length(client *http.Client, baseURL string) int64 {
    resp, err := client.Get(baseURL + "/a-random-string-that-does-not-exist-123")
    if err != nil { return -1 }
    defer resp.Body.Close()
    return resp.ContentLength
}



func ConcurrentScan(ctx context.Context, urlTemplate string, wordlist []string, workerCount int, filterCodes string, recursive bool, maxDepth int, delay time.Duration, headers map[string]string) <-chan ScanResult {

	results := make(chan ScanResult, 500)

	filterMap := make(map[int]bool)

	for _, codeStr := range strings.Split(filterCodes, ",") {
		if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil { filterMap[code] = true }
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}

	badContentLength := get404Length(client, urlTemplate)

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
						ticker = time.NewTicker(delay)
						defer ticker.Stop()
					}
					for target := range jobs {
						select {
						case <-ctx.Done():
							return
						default:
							if ticker != nil { <-ticker.C }
						}
						if _, loaded := globalSeen.LoadOrStore(target, true); loaded { continue }

						req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
						if err != nil { continue }

						for key, value := range headers { req.Header.Set(key, value) }

						resp, err := client.Do(req)
						if err != nil { continue }

						body, _ := io.ReadAll(resp.Body)
						resp.Body.Close()
						currentLen := int64(len(body))

						status := resp.StatusCode
						loc := resp.Header.Get("Location")

						if !filterMap[status] {
							if status == 200 && currentLen == badContentLength {
                                                                continue // It's a Soft 404, ignore it
                                                        }
							results <- ScanResult{
								URL: target, 
								StatusCode: status, 
								Location: loc,
								ContentLength: currentLen,
							}

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
