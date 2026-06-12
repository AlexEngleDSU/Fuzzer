package engine
import (
	"context"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
	"fmt"
	fhttp "github.com/bogdanfinn/fhttp"
)


func ConcurrentScan(ctx context.Context, host, urlTemplate, headerTemplate string, wordlist []string, workerCount int, filterCodes string, recursive bool, maxDepth int, delay time.Duration, onStatusUpdate func(string),) <-chan ScanResult {
	results := make(chan ScanResult, 500)
	jobs := make(chan Job, 1000)
	discovery := make(chan Job, 1000)
	filterMap := make(map[int]bool)
	
	queue := []Job{{URL: urlTemplate, Depth: 0}}
	
	browserClient := CreateBrowserClient()
	baseHeaders := fmt.Sprintf(headerTemplate, host, host)
	headers := GetOrderedHeaders(baseHeaders)
	badContentLength := get404Length(browserClient, urlTemplate)
	onStatusUpdate("Starting Scan!")
    	time.Sleep(1 * time.Second)
	
	for _, codeStr := range strings.Split(filterCodes, ",") {
		if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil { filterMap[code] = true }
	}
	onStatusUpdate("Starting Scan!")
	time.Sleep(1 * time.Second)
	onStatusUpdate("")
	go func() {
		defer close(results)
		defer close(jobs)
		defer close(discovery)		
		
		var wg sync.WaitGroup
		globalSeen := sync.Map{}
		
		for w := 1; w <= workerCount; w++ {
			go func() {
				for job := range jobs {
					func(j Job) {
						defer wg.Done()
						PauseMu.Lock()
						if Paused {
							pChan := PauseChan
							PauseMu.Unlock()
							<- pChan
						} else { PauseMu.Unlock() }
						if delay > 0 {
							select {
								case <- time.After(delay):
								case <- ctx.Done(): return
							}
						}
						if _, loaded := globalSeen.LoadOrStore(j.URL, true); loaded {
							return
						}
						req, err := fhttp.NewRequest("GET", j.URL, nil)
						if err != nil {
							return
						}
						for _, h := range headers {
							if strings.EqualFold(h.Key, "Host") {
								req.Host = h.Value
								continue
							}
							req.Header.Add(h.Key, h.Value)
						}
						resp, err := browserClient.Do(req)
						if err != nil {
							return
						}
						body, _ := io.ReadAll(resp.Body)
						resp.Body.Close()
						respCookies := resp.Cookies()
						status := resp.StatusCode
						if !filterMap[status] && !(status == 200 && int64(len(body)) == badContentLength) {
							results <- ScanResult{
								URL:	j.URL,
								StatusCode:    status,
								ContentLength: int64(len(body)),
								Location:      resp.Header.Get("Location"),
								Cookies:       respCookies,
								Depth:         j.Depth,
							}
							onStatusUpdate(fmt.Sprintf("--- Starting Depth %d | Processing %d seeds ---", j.Depth, len(queue)))
							if recursive && status >= 300 && status < 400 {
								resolved := ResolveURL(j.URL, resp.Header.Get("Location"))
								onStatusUpdate(fmt.Sprintf("Discovered new path at depth %d: %s", j.Depth + 1, resolved))
								discovery <- Job{URL: resolved, Depth: j.Depth + 1}
							}
						}
					} (job)
				}
			}()
		}
		for depth := 0; depth <= maxDepth; depth++ {
			onStatusUpdate(fmt.Sprintf("Scanning depth %d...", depth))
			for _, base := range queue {
				for _, word := range wordlist {
					target := strings.ReplaceAll(base.URL, "FUZZ", word)
					wg.Add(1)
					jobs <- Job{URL: target, Depth: base.Depth}
				}
			}
			wg.Wait()
			
			newQueue := []Job{}
			for len(discovery) > 0 {
				d := <-discovery
				if !strings.Contains(d.URL, "FUZZ") {
					if !strings.HasSuffix(d.URL, "/") { d.URL += "/" }
					d.URL += "FUZZ"
				}
				if _, seen := globalSeen.LoadOrStore(d, true); !seen && d.Depth <= maxDepth {
					newQueue = append(newQueue, d)
				}
			}
			queue = newQueue
			if len(queue) == 0 { break }
		}
	}()
	return results
}

