package cli

import (
    "context"
    "flag"
    "fmt"
    "time"
    "strings"
    "os"
    "bufio"
    "net/url"
    "github.com/AlexEngleDSU/Fuzzer/pkg/engine"
    "github.com/AlexEngleDSU/Fuzzer/pkg/appUI"
    "github.com/AlexEngleDSU/Fuzzer/pkg/browser"
)

func loadWordlist(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        // Use TrimSpace to clean up whitespace and ensure it's a string
        line := strings.TrimSpace(scanner.Text())
        
        // Now 'line' is a string, so this check works perfectly
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        lines = append(lines, line)
    }
    return lines, scanner.Err()
}

func Run(args []string) {
    fs := flag.NewFlagSet("cli", flag.ExitOnError)
    
    urlInput := fs.String("u", "", "Target URL")
    wordlistPath := fs.String("w", "", "Path to wordlist file")
    threads := fs.Int("T", 10, "threads")
    depth := fs.Int("r", 0, "recursion depth")
    delay := fs.Int("d", 1, "delay time")
    timeout := fs.Int("t", 0, "timeout")	
    matchCodes := fs.String("m", "", "Match Status Codes")
    filterCodes := fs.String("f", "", "Filter Status Codes")
    verbose := fs.Bool("v", false, "Enable verbose output")
    fs.Parse(args)
    
    wordlist, err := loadWordlist(*wordlistPath)
    if err != nil {
    	fmt.Printf("Error loading wordlist: %v\n", err)
    	return
    }
    
    baseURL := strings.Replace(*urlInput, "/FUZZ", "/", 1)
    
    u, err := url.Parse(baseURL)
    if err != nil {
    	fmt.Println("Invalid URL")
    	return
    }
    host := u.Host

    // CLI-specific status handler
    onStatus := func(msg string) {
    	if *verbose {
    		fmt.Printf("[STATUS] %s\n", msg)
    	}
    }
    
    fmt.Println("[+] Initializing browser session to bypass WAF...\n")
    session, err := browser.InitializeSession(baseURL)
    if err != nil { fmt.Printf("[-] Browser Error: %v\n", err) }
    
    if session != nil && len(session.Cookies) > 0 {
	    convertedCookies := make([]browser.Cookie, 0, len(session.Cookies))
	    for _, c := range session.Cookies {
		name, _ := c["name"].(string)
		value, _ := c["value"].(string)
		if name != "" {
		    convertedCookies = append(convertedCookies, browser.Cookie{
		        Name:  name,
		        Value: value,
		    })
		}
	    }
            browser.InjectCookies(engine.GlobalJar, *urlInput, convertedCookies)
            allCookiesMap := engine.GlobalJar.GetAllCookies()
            for _ , cookies := range allCookiesMap {
		for _, c := range cookies {
			fmt.Printf("%s: %s\n\n", c.Name, c.Value)
		}
	    }
    } else { 
        fmt.Println("[-] Skipping cookie injection: No cookies found in session.") 
    }
	
    formattedHeaders := fmt.Sprintf(appUI.HeaderTemplate, host, baseURL)
    
    results := engine.ConcurrentScan(
    	context.Background(),
    	host, // host
    	*urlInput, //urlTemplate
    	formattedHeaders, // headerTemplate
    	wordlist, // wordlist (load from file in a real app)
    	true,
    	*depth,
    	*threads,
    	time.Duration(*delay) * time.Second,
    	time.Duration(*timeout) * time.Second,
    	*matchCodes,
    	*filterCodes,
    	onStatus,
    )

    for res := range results {
    	if res.StatusCode == 301 || res.StatusCode == 302 { fmt.Printf("[%d] %s -> %s\n", res.StatusCode, res.URL, res.Location) 
    	} else { fmt.Printf("[%d] <%d> %s\n", res.StatusCode, res.ContentLength, res.URL) }
    }    
}

