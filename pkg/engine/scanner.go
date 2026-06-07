package engine

import (
    "bufio"
    "context"
    "fmt"
    "io"
    "net/url"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"

    fhttp "github.com/bogdanfinn/fhttp"
    "github.com/bogdanfinn/fhttp/cookiejar" 
    tls_client "github.com/bogdanfinn/tls-client"
    "github.com/bogdanfinn/tls-client/profiles"
)

type ScanResult struct {
	URL           string
	StatusCode    int
	Location      string
	Message       string
	ContentLength int64
	Cookies       []*fhttp.Cookie
}

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

func resolveURL(base, loc string) string {
	if strings.HasPrefix(loc, "http") {
		return loc
	}
	if strings.HasPrefix(loc, "//") {
		return "https:" + loc
	}
	parsedBase, _ := url.Parse(base)
	cleanLoc := strings.TrimLeft(loc, "/")
	return fmt.Sprintf("%s://%s/%s", parsedBase.Scheme, parsedBase.Host, cleanLoc)
}

// 1. Change the signature to match your browserClient type
func get404Length(client tls_client.HttpClient, baseURL string) int64 {
    req, _ := fhttp.NewRequest("GET", baseURL + "/a-random-string-that-does-not-exist-123", nil)
    
    // 2. Use the 'client' argument passed to the function
    resp, err := client.Do(req)
    if err != nil {
        return -1
    }
    defer resp.Body.Close()
    return resp.ContentLength
}

// 3. Change the signature here as well
func performHandshake(client tls_client.HttpClient, host string) {
    target := host
    if !strings.HasPrefix(target, "http") { target = "https://" + target }
    
    req, _ := fhttp.NewRequest("GET", target, nil)
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("Handshake Error: %v\n", err)
        return
    }
    
    // Check if the server sent ANY cookies
    cookies := resp.Cookies()
    fmt.Printf("DEBUG: Handshake received %d cookies: %+v\n", len(cookies), cookies)
    
    resp.Body.Close()
}

var (
	paused    bool
	pauseMu   sync.Mutex
	pauseChan = make(chan struct{})
)

func SetPause(p bool) {
	pauseMu.Lock()
	defer pauseMu.Unlock()
	if p != paused {
		paused = p
		if !paused {
			close(pauseChan)
			pauseChan = make(chan struct{})
		}
	}
}

type headerTransport struct {
    headers   map[string]string
    transport fhttp.RoundTripper // Add this field
}

func (t *headerTransport) RoundTrip(req *fhttp.Request) (*fhttp.Response, error) {
    req2 := req.Clone(req.Context())
    req2.Header.Del("User-Agent")
    
    // Use the transport we passed in, NOT DefaultTransport
    return t.transport.RoundTrip(req2)
}

func getBaseURL(rawURL string) string {
    u, err := url.Parse(rawURL)
    if err != nil {
        return "/"
    }
    // Reconstruct the directory base: Scheme + Host + Path up to last slash
    base := u.Scheme + "://" + u.Host + u.Path
    lastSlash := strings.LastIndex(base, "/")
    if lastSlash != -1 {
        return base[:lastSlash+1]
    }
    return base + "/"
}

func CreateBrowserClient() tls_client.HttpClient {
    // This profile mimics Chrome 146's TLS handshake
    jar, _ := cookiejar.New(&cookiejar.Options{})
    
//    proxyUrl := "http://127.0.0.1:8080"
    
    options := []tls_client.HttpClientOption{
        tls_client.WithTimeoutSeconds(30),
        tls_client.WithClientProfile(profiles.Chrome_146), // Matches your browser
        tls_client.WithCookieJar(jar),
// 	Disable Redirects
        tls_client.WithNotFollowRedirects(),
//	Burp
//      tls_client.WithProxyUrl(proxyUrl),
//      tls_client.WithInsecureSkipVerify(),
    }

    client, _ := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
    return client
}

type HeaderLine struct {
    Key   string
    Value string
}

// Parse your template into an ordered slice
func GetOrderedHeaders(template string) []HeaderLine {
    var ordered []HeaderLine
    lines := strings.Split(strings.ReplaceAll(template, "\r", ""), "\n")
    
    for _, line := range lines {
        if strings.Contains(line, ":") && !strings.Contains(line, "HTTP/") {
            parts := strings.SplitN(line, ":", 2)
            ordered = append(ordered, HeaderLine{
                Key:   strings.TrimSpace(parts[0]),
                Value: strings.TrimSpace(parts[1]),
            })
        }
    }
    // Print the local variable!
    fmt.Printf("DEBUG: Successfully parsed %d headers\n", len(ordered))
    return ordered
}

func ConcurrentScan(ctx context.Context, host, urlTemplate, headerTemplate string, wordlist []string, workerCount int, filterCodes string, recursive bool, maxDepth int, delay time.Duration) <-chan ScanResult {
	
	results := make(chan ScanResult, 500)
	
	browserClient := CreateBrowserClient() 
	
	resp, err := browserClient.Get(getBaseURL(urlTemplate))
	if err == nil {
		resp.Body.Close()
	}
	
	
        fmt.Printf("DEBUG: Header Template received: '%s'\n", headerTemplate)
	headers := GetOrderedHeaders(headerTemplate)
	fmt.Printf("DEBUG: Headers count: %d\n", len(headers))
        lastRedirectURL := &sync.Map{}
        lastRedirectURL.Store("latest", getBaseURL(urlTemplate))
    
    // 2. Perform initial handshake
        _, _ = browserClient.Get(getBaseURL(urlTemplate))
	
	filterMap := make(map[int]bool)

	for _, codeStr := range strings.Split(filterCodes, ",") {
		if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil {
			filterMap[code] = true
		}
	}

	// Perform Handshake
	performHandshake(browserClient, host)

	badContentLength := get404Length(browserClient, urlTemplate)
	
	var closeOnce sync.Once
	

	go func() {	
	
		closeChannel := func() {
			closeOnce.Do(func() {
	    			close(results)
			})
		}
		
		defer closeChannel()
		
		queue := []string{urlTemplate}

		for depth := 0; depth <= maxDepth; depth++ {
			if len(queue) == 0 {
				break
			}

			results <- ScanResult{Message: fmt.Sprintf("[+] Starting depth level: %d | Base Paths: %d", depth, len(queue))}

			nextGen := []string{}
			var mu sync.Mutex
			var globalSeen sync.Map
			var wg sync.WaitGroup
			jobs := make(chan string, 100)
			var once sync.Once

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
						pauseMu.Lock()
						if paused {
							pChan := pauseChan
							pauseMu.Unlock()
							<-pChan
						} else {
							pauseMu.Unlock()
						}
						
						select {
						case <-ctx.Done():
							return
						default:
							if ticker != nil {
								<-ticker.C
							}
						}

						if _, loaded := globalSeen.LoadOrStore(target, true); loaded {
							continue
						}

						req, err := fhttp.NewRequest("GET", target, nil)
						if err != nil {
							continue
						}
						
						req.Header = fhttp.Header{}
						
						for _, h := range headers {
						    // Host is special in Go's fhttp/net/http; handle it directly
						    if strings.EqualFold(h.Key, "Host") {
						        req.Host = h.Value
						        continue
						    }
						    req.Header.Add(h.Key, h.Value)
						}
						
						if val, ok := lastRedirectURL.Load("latest"); ok {
					                req.Header.Set("Referer", val.(string))
				                }
				                
				                targetURL, _ := url.Parse(target)
					        cookies := browserClient.GetCookieJar().Cookies(targetURL)
					        fmt.Printf("DEBUG: Sending %d cookies: %+v\n", len(cookies), cookies)
					        
					        req.Header.Set("Connection", "keep-alive")
						req.Header.Set("Upgrade-Insecure-Requests", "1")
						req.Header.Set("Sec-Fetch-Site", "same-origin")
						
						resp, err := browserClient.Do(req)
						
						once.Do(func() {
							fmt.Printf("--- FIRST REQUEST DUMP ---\n%s\n", req.Header)
							fmt.Printf("--- RESPONSE DUMP ---\n%s\n", resp.StatusCode)
							
						})

						
						if err != nil {
							continue
						}
						
						body, _ := io.ReadAll(resp.Body)
						resp.Body.Close()
						currentLen := int64(len(body))
						status := resp.StatusCode
						loc := resp.Header.Get("Location")
						
						resolved := ""
						if loc != "" {
						    resolved = resolveURL(target, loc)
						    
						    // Store for Referer tracking
						    lastSlash := strings.LastIndex(resolved, "/")
						    if lastSlash != -1 {
							cleanReferer := resolved[:lastSlash+1]
							lastRedirectURL.Store("latest", cleanReferer)
						    }
						}

						if !filterMap[status] {
						    if status == 200 && currentLen == badContentLength {
							continue
						    }
						    
						    displayURL := target
						    if status >= 300 && status < 400 && resolved != "" {
								displayURL = fmt.Sprintf("%s -> %s", target, resolved)
	     					    }						    

						    results <- ScanResult{
							URL:           displayURL,
							StatusCode:    status,
							Location:      loc, // loc is the raw header, you might prefer 'resolved' here
							ContentLength: currentLen,
							Cookies:       cookies,
						    }

						    // 2. Reuse the already resolved URL
						    if recursive && (status >= 300 && status < 400) && resolved != "" {
							//newPath := strings.TrimSuffix(resolved, "/") + "/FUZZ"
							mu.Lock()
							nextGen = append(nextGen, resolved)
							mu.Unlock()
						    }
						}
					}
				}()
			}

			for _, base := range queue {
				for _, word := range wordlist {
					finalURL := strings.ReplaceAll(base, "FUZZ", word)
					jobs <- finalURL
				}
			}
			close(jobs)
			wg.Wait()
			queue = nextGen
		}
	}()
	return results
}
