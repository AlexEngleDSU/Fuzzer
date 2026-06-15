package engine

import (
	"bufio"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"encoding/json"
	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type Job struct {
	URL   string
	Depth int
}
// GlobalJar initialized with options to prevent compilation errors
var GlobalJar = tls_client.NewCookieJar()

type Pauser struct {
    mu      sync.Mutex
    cond    *sync.Cond
    paused  bool
}

func NewPauser() *Pauser {
    p := &Pauser{paused: false}
    p.cond = sync.NewCond(&p.mu)
    return p
}

func (p *Pauser) Wait() {
    p.mu.Lock()
    defer p.mu.Unlock()
    for p.paused { p.cond.Wait() }
}

func (p *Pauser) SetPause(paused bool) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.paused = paused
    if !paused { p.cond.Broadcast() } // Wake up all waiting workers
}

type Config struct {
    LastWordlist string `json:"last_wordlist"`
    LastDir      string `json:"last_dir"`
}

func SaveLastFilePath(path string) {
    config := Config{
        LastWordlist: path,
        LastDir:      filepath.Dir(path),
    }
    data, _ := json.MarshalIndent(config, "", "  ")
    os.WriteFile(GetConfigPath(), data, 0644)
}

func GetLastFilePath() string {
    data, err := os.ReadFile(GetConfigPath())
    if err != nil { return "" }
    var config Config
    json.Unmarshal(data, &config)
    return config.LastWordlist
}

func GetInitialPath() string {
    data, err := os.ReadFile(GetConfigPath())
    if err != nil { return "" }
    var config Config
    json.Unmarshal(data, &config)
    return config.LastDir
}

func GetConfigPath() string { // Capitalized to be exported
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	appDir := filepath.Join(configDir, "fuzzer")
	os.MkdirAll(appDir, 0700)
	return filepath.Join(appDir, "config.json")
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

func ResolveURL(base, loc string) string {
	if strings.HasPrefix(loc, "http") { return loc }
	u, err := url.Parse(base)
	if err != nil { return loc }
	rel, err := url.Parse(loc)
	if err != nil { return loc }
	return u.ResolveReference(rel).String()
}

func GetOrderedHeaders(template string) []HeaderLine {
	var ordered []HeaderLine
	scanner := bufio.NewScanner(strings.NewReader(template))
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, ":"); idx != -1 {
			ordered = append(ordered, HeaderLine{
				Key:   strings.TrimSpace(line[:idx]),
				Value: strings.TrimSpace(line[idx+1:]),
			})
		}
	}
	return ordered
}

func get404Length(client tls_client.HttpClient, baseURL string) int64 {
	req, _ := fhttp.NewRequest("GET", baseURL+"/a-random-string-that-does-not-exist-123", nil)
	resp, err := client.Do(req)
	if err != nil { return -1 }
	defer resp.Body.Close()
	return resp.ContentLength
}

func CreateBrowserClient(timeout time.Duration, options ScanOptions) tls_client.HttpClient {
    timeoutSecs := int(timeout.Seconds())
    
    // 1. Determine which jar to use
    var jar tls_client.CookieJar
    if options.Stateful {
        jar = GlobalJar
    } else {
        // cookiejar.New requires an options struct, use &cookiejar.Options{}
        jar = tls_client.NewCookieJar()
    }
    
    // 2. Pass the 'jar' variable into the options
    clientOptions := []tls_client.HttpClientOption{
        tls_client.WithTimeoutSeconds(timeoutSecs),
        tls_client.WithClientProfile(profiles.Chrome_146),
        tls_client.WithCookieJar(jar), // USE THE LOCAL 'jar' VARIABLE HERE
        tls_client.WithNotFollowRedirects(),
    }
    
    client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), clientOptions...)
    if err != nil { return nil }
    return client
}

func FilterCookieJar(jar fhttp.CookieJar, allowList []string) {
	// 1. Create a map for O(1) lookup
	allowed := make(map[string]bool)
	for _, name := range allowList {
		allowed[name] = true
	}

	// 2. Iterate through all cookies in the jar
	// Note: We need a URL to retrieve cookies from the jar
	// If you are fuzzing one host, we can extract it from the jar's cookies
	// or you can pass the target URL here.
	
	// Since we don't have a single URL, we have to collect all cookies
	// and re-set them. This logic assumes your GlobalJar is accessible.
	
	// If you are using bogdanfinn/tls-client, it often provides a way 
	// to manipulate the jar directly.
	
	// General approach:
	allCookies := GlobalJar.GetAllCookies() // Assuming your GlobalJar has this method
	
	for uStr, cookies := range allCookies {
		u, _ := url.Parse(uStr)
		var filtered []*fhttp.Cookie
		
		for _, c := range cookies {
			if allowed[c.Name] {
				filtered = append(filtered, c)
			}
		}
		// Set the filtered list back to the jar
		GlobalJar.SetCookies(u, filtered)
	}
}




