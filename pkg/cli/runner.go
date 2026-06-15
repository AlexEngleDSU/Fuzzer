package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/AlexEngleDSU/Fuzzer/pkg/appUI"
	"github.com/AlexEngleDSU/Fuzzer/pkg/browser"
	"github.com/AlexEngleDSU/Fuzzer/pkg/engine"
	"github.com/bogdanfinn/tls-client"
)

func Run(args []string) {
	fs := flag.NewFlagSet("cli", flag.ExitOnError)

	// Flags
	browsingState := fs.Bool("S", false, "Enable Stateful Fuzzing")
	cookieAllowList := fs.String("cal", "", "Comma-separated Cookie Allow List")
	urlInput := fs.String("u", "", "Target URL")
	wordlistPath := fs.String("w", "", "Path to wordlist file")
	threads := fs.Int("T", 10, "Threads")
	depth := fs.Int("r", 0, "Recursion depth")
	delay := fs.Int("d", 1, "Delay (seconds)")
	timeout := fs.Int("t", 5, "Timeout (seconds)")
	matchCodes := fs.String("m", "", "Match Status Codes")
	filterCodes := fs.String("f", "", "Filter Status Codes")
	verbose := fs.Bool("v", false, "Enable verbose output")

	fs.Parse(args)

	// 1. Prepare Configuration
	allowList := parseAllowList(*cookieAllowList)
	options := engine.ScanOptions{
		Stateful:        *browsingState,
		Timeout:         time.Duration(*timeout) * time.Second,
		CookieAllowList: allowList,
	}

	// 2. Load Resources
	wordlist, err := loadWordlist(*wordlistPath)
	if err != nil {
		fmt.Printf("[-] Error loading wordlist: %v\n", err)
		return
	}

	baseURL := strings.Replace(*urlInput, "/FUZZ", "/", 1)
	u, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("[-] Invalid URL")
		return
	}

	// 3. Handle Browser/State
	handleSession(*browsingState, baseURL, *urlInput, allowList)

	// 4. Execute Scan
	onStatus := func(msg string) {
		if *verbose { fmt.Printf("[STATUS] %s\n", msg) }
	}

	results := engine.ConcurrentScan(
		options,
		context.Background(),
		u.Host,
		*urlInput,
		fmt.Sprintf(appUI.HeaderTemplate, u.Host, baseURL),
		wordlist,
		*depth > 0,
		*depth,
		*threads,
		time.Duration(*delay)*time.Second,
		options.Timeout,
		*matchCodes,
		*filterCodes,
		onStatus,
	)

	// 5. Output Results
	for res := range results {
		if res.StatusCode >= 300 && res.StatusCode < 400 {
			fmt.Printf("[%d] %s -> %s\n", res.StatusCode, res.URL, res.Location)
		} else {
			fmt.Printf("[%d] <%d> %s\n", res.StatusCode, res.ContentLength, res.URL)
		}
	}
}

// Helpers for clean code
func parseAllowList(input string) []string {
	if input == "" { return nil }
	parts := strings.Split(input, ",")
	for i := range parts { parts[i] = strings.TrimSpace(parts[i]) }
	return parts
}

func handleSession(stateful bool, baseURL, fullURL string, allowList []string) {
	if !stateful {
		engine.GlobalJar = tls_client.NewCookieJar()
		fmt.Println("[-] Non-stateful mode: Session cleared.")
		return
	}

	fmt.Println("[+] Stateful mode: Initializing browser...")
	session, err := browser.InitializeSession(baseURL)
	if err != nil {
		fmt.Printf("[-] Browser Error: %v\n", err)
		return
	}

	if session != nil && len(session.Cookies) > 0 {
		var cookies []browser.Cookie
		for _, c := range session.Cookies {
			if name, _ := c["name"].(string); name != "" {
				cookies = append(cookies, browser.Cookie{Name: name, Value: c["value"].(string)})
			}
		}
		browser.InjectCookies(engine.GlobalJar, fullURL, cookies)
		allCookiesMap := engine.GlobalJar.GetAllCookies()
	        for _ , cookies := range allCookiesMap {
		    for _, c := range cookies {
			fmt.Printf("%s: %s\n", c.Name, c.Value)
		    }
	        }
		if len(allowList) > 0 {
			engine.FilterCookieJar(engine.GlobalJar, allowList)
		}
		fmt.Println("[+] Session established and cookies injected.")
	}
}

func loadWordlist(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil { return nil, err }
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") { continue }
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}
