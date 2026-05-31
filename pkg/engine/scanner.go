package engine

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

type ScanResult struct {
	URL        string
	StatusCode int
	Location   string
}

func ConcurrentScan(urlTemplate string, wordlist []string, workerCount int, rps int, quiet bool, outputFile string, filterCodes string) {
	var f *os.File
	if outputFile != "" {
		f, _ = os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			defer f.Close()
		}
	}

	filterMap := make(map[int]bool)
	for _, codeStr := range strings.Split(filterCodes, ",") {
		if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil {
			filterMap[code] = true
		}
	}

	results := make(chan ScanResult, 100)
	jobs := make(chan string, 100)
	var wg sync.WaitGroup

	client := &http.Client{
	    Timeout: 5 * time.Second,
	    // This forces the client to return the 30x response instead of following it
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
				// Inject the word into the template
				targetURL := strings.ReplaceAll(urlTemplate, "FUZZ", word)

				resp, err := client.Get(targetURL)
				if err != nil { continue }

				loc := resp.Header.Get("Location")

				// Efficiency Filter: Drop 404s early if quiet
				if quiet && resp.StatusCode == 404 {
					resp.Body.Close()
					continue
				}

				if !filterMap[resp.StatusCode] {
					results <- ScanResult{URL: targetURL, StatusCode: resp.StatusCode, Location: loc}
				}
				resp.Body.Close()
			}
		}()
	}

	// 2. Orchestrator
	go func() {
		for res := range results {
			msg := formatResult(res)
			fmt.Println(msg)
			if f != nil { f.WriteString(msg + "\n") }
		}
	} ()

	// 3. Job Dispatcher
	for _, word := range wordlist {
		jobs <- word
	}
	close(jobs)
	wg.Wait()
	close(results)
}

func formatResult(res ScanResult) string {
	switch res.StatusCode {
	case 200:
		return fmt.Sprintf("[+] Found: %s (Status: 200)", res.URL)
	case 301, 302:
		return fmt.Sprintf("[>] Redirect: %s (Status: %d)", res.URL, res.Location, res.StatusCode)
	case 401, 403:
		return fmt.Sprintf("[X] Unauthorized: %s (Status: %d)", res.URL, res.StatusCode)
	case 404:
		return fmt.Sprintf("[-] Not Found: %s (Status: 404)", res.URL)
	default:
		return fmt.Sprintf("[?] Unknown: %s (Status: %d)", res.URL, res.StatusCode)
	}
}
