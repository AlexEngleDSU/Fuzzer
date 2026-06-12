package main

import (
    "context"
    "flag"
    "fmt"
    "time"
    "os"
    "bufio"
    "net/url"
    "github.com/AlexEngleDSU/Fuzzer/pkg/engine"
    "github.com/AlexEngleDSU/Fuzzer/pkg/ui"
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
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}

func main() {
    urlInput := flag.String("u", "", "Target URL")
    wordlistPath := flag.String("w", "", "Path to wordlist file")
    threads := flag.Int("t", 10, "threads")
    filter := flag.String("f", "", "Filter")
    depth := flag.Int("r", 0, "recursion depth")
    delay := flag.Int("d", 1, "delay time")
    verbose := flag.Bool("v", false, "Enable verbose output")
    flag.Parse()
    
    wordlist, err := loadWordlist(*wordlistPath)
    if err != nil {
    	fmt.Printf("Error loading wordlist: %v\n", err)
    	return
    }
    
    u, err := url.Parse(*urlInput)
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

    results := engine.ConcurrentScan(
    	context.Background(),
    	host, // host
    	*urlInput, //urlTemplate
    	ui.HeaderTemplate, // headerTemplate
    	wordlist, // wordlist (load from file in a real app)
    	*threads,
    	*filter,
    	true,
    	*depth,
    	time.Duration(*delay) * time.Second,
    	onStatus,
    )

    for res := range results {
        fmt.Printf("[%d] %s\n", res.StatusCode, res.URL)
    }
}
