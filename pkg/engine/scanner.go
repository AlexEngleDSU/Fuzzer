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
func ConcurrentScan(baseURL string, paths []string, workerCount int, rps int, quiet bool) {

	delay := time.Second / time.Duration(rps)
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

        jobs := make(chan string, len(paths))
        results := make(chan string)
        var wg sync.WaitGroup

        client := &http.Client{
    		Timeout: 5 * time.Second,
    		CheckRedirect: func(req *http.Request, via []*http.Request) error {
        		// This stops the client from following redirects,
        		// allowing your code to see the 301/302 status codes.
        		return http.ErrUseLastResponse
		},
	}
	
        // 1. Start workers
        for w := 1; w <= workerCount; w++ {
                wg.Add(1)
                go func() {
                        defer wg.Done()
                        for path := range jobs {
                        	<-ticker.C
                                url := fmt.Sprintf("%s/%s", baseURL, path)
                                
                                resp, err := client.Get(url)
                                if err != nil {
                                	continue
                                }

                                resp.Body.Close()

				length := resp.ContentLength
				sizeStr := fmt.Sprintf("%d", length) // Convert int64 to string
				if length == -1 {
    					sizeStr = "unknown"
				}

                                switch resp.StatusCode {
				case 200:
				    results <- fmt.Sprintf("[+] Found: %s (Status: 200) [Size: %s]", url, sizeStr)
				case 301, 302:
				    loc := resp.Header.Get("Location")
				    results <- fmt.Sprintf("[>] Redirect: %s -> %s (Status: %d) [Size: %s]", url, loc, resp.StatusCode, sizeStr)
				case 401, 403:
				    results <- fmt.Sprintf("[X] Unauthorized: %s (Status: %d) [Size: %s]", url, resp.StatusCode, sizeStr)
				case 404:
					if !quiet {
				    		results <- fmt.Sprintf("[-] Not Found: %s (Status: 404) [Size: %s]", url, sizeStr)
				    	}
				default:
				    // This catches other codes like 500 (Internal Server Error)
				    results <- fmt.Sprintf("[?] Unknown: %s (Status: %d) [Size: %s]", url, resp.StatusCode, sizeStr)
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
