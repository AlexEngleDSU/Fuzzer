package engine

import (
        "bufio"
        "fmt"
        "net/http"
        "os"
        "sync"
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
func ConcurrentScan(baseURL string, paths []string, workerCount int) {
        jobs := make(chan string, len(paths))
        results := make(chan string)
        var wg sync.WaitGroup

        // 1. Start workers
        for w := 1; w <= workerCount; w++ {
                wg.Add(1)
                go func() {
                        defer wg.Done()
                        for path := range jobs {
                                url := fmt.Sprint("%s/%s", baseURL, path)
                                resp, err := http.Get(url)
                                if err == nil && resp.StatusCode == 200 {
                                        results <- fmt.Sprintf("[+] Found: %s (Status: 200)", url)
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
