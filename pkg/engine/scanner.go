package engine

import (
        "bufio"
        "fmt"
        "net/http"
        "os"
        "sync"
        "time"
)

// ReadLines: Helper to read wordlist files
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

// ConcurrentScan: manage workers to scan paths in parallel
func ConcurrentScan(baseURL string, paths []string, workerCount int, rps int) {

	delay := time.Second / time.Duration(rps)
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

        jobs := make(chan string, len(paths))
        results := make(chan string)
        var wg sync.WaitGroup

        // 1. Start workers
        for w := 1; w <= workerCount; w++ {
                wg.Add(1)
                go func() {
                        defer wg.Done()
                        for path := range jobs {
                        	<-ticker.C
                                url := fmt.Sprintf("%s/%s", baseURL, path)
                                resp, err := http.Get(url)
                                if err != nil {
                                	continue
                                }
                                defer resp.Body.Close()

                                switch resp.StatusCode {
				case 200:
				    results <- fmt.Sprintf("[+] Found: %s (Status: 200)", url)
				case 301, 302:
				    results <- fmt.Sprintf("[>] Redirect: %s (Status: %d)", url, resp.StatusCode)
				case 401, 403:
				    results <- fmt.Sprintf("[X] Unauthorized: %s (Status: %d)", url, resp.StatusCode)
				case 404:
				    results <- fmt.Sprintf("[-] Not Found: %s (Status: 404)", url)
				default:
				    // This catches other codes like 500 (Internal Server Error)
				    results <- fmt.Sprintf("[?] Unknown: %s (Status: %d)", url, resp.StatusCode)
				}
                        }
                }()
        }

        // 2. Populate workers
        go func() {
                for _, path := range paths {
                        jobs <- path
                }
                close(jobs)
        }()

        // 3. Wait / Close results
        go func() {
                wg.Wait()
                close(results)
        }()

        // 4. Print results in real time
        for res := range results {
                fmt.Println(res)
        }
}
