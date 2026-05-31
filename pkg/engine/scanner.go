package engine

import (
	"bufio"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ScanResult holds the data for the UI to consume
type ScanResult struct {
	URL        string
	StatusCode int
	Location   string
}

// ReadLines helper to load wordlists efficiently
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// ConcurrentScan is now decoupled: it returns a channel for the UI to listen to.
func ConcurrentScan(urlTemplate string, wordlist []string, workerCount int, filterCodes string) chan ScanResult {
	results := make(chan ScanResult, 100)
	jobs := make(chan string, 100)
	var wg sync.WaitGroup

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

	// 1. Worker Pool
	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for word := range jobs {
				targetURL := strings.ReplaceAll(urlTemplate, "FUZZ", word)
				
				resp, err := client.Get(targetURL)
				if err != nil { continue }

				loc := resp.Header.Get("Location")
				statusCode := resp.StatusCode
				resp.Body.Close()

				if !filterMap[statusCode] {
					results <- ScanResult{URL: targetURL, StatusCode: statusCode, Location: loc}
				}
			}
		}()
	}

	// 2. Job Dispatcher
	go func() {
		for _, word := range wordlist {
			jobs <- word
		}
		close(jobs)
	}()

	// 3. Closer: Ensures channel closes only after workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
