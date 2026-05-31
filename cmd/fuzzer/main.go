package main

import (
        "fmt"
        "os"
        "github.com/AlexEngleDSU/Fuzzer/pkg/engine"
        "github.com/spf13/cobra"
)

var (
        targetURL string
        threadCount int
        wordlistPath string
        rate int
        quiet bool
        outputFile string
        filterCodes string
        maxDepth int = 1
)

var rootCmd = &cobra.Command{
        Use: "fuzzer",
        Short: "Professional fuzzer",
        Run: func(cmd *cobra.Command, args []string) {
                if targetURL == "" || wordlistPath == "" {
                        fmt.Println("[-] Error: -u and -w are required")
                        return
                }

                paths, err := engine.ReadLines(wordlistPath)
                if err != nil {
                        fmt.Printf("[-] Error: wordlist: %v\n", err)
                        return
                }

                fmt.Println("-------------------------------------------")
                fmt.Printf("[+] Scanning target: %s\n", targetURL)
                fmt.Printf("[+] Wordlist: %d entries from %s\n", len(paths), wordlistPath)
                fmt.Printf("[+] Thread count: %d\n", threadCount)
                fmt.Printf("[+] Rate count: %d\n", rate)
                fmt.Printf("[+] Output file: %s\n", outputFile)
                fmt.Printf("[-] Filtering codes: %s\n", filterCodes)
                fmt.Printf("[+] Recursion depth: %d\n", maxDepth)
                fmt.Println("-------------------------------------------")
                engine.ConcurrentScan(targetURL, paths, threadCount, rate, quiet, outputFile, filterCodes)
        },
}

func init() {
        // Define flags
        rootCmd.Flags().StringVarP(&targetURL, "url", "u", "", "URL to scan")
        rootCmd.Flags().StringVarP(&wordlistPath, "wordlist", "w", "", "Wordlist selected")
        rootCmd.Flags().IntVarP(&threadCount, "threads", "t", 10, "Number of concurrent threads")
        rootCmd.Flags().IntVarP(&rate, "rate", "r", 10, "Number of requests per second")
        rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only show found/interesting results")
        rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file to save results")
        rootCmd.Flags().StringVar(&filterCodes, "filter", "", "Comma-separated status codes to ignore")
    
	// 2. Add the alias exactly as before
	rootCmd.Flags().StringVar(&filterCodes, "fc", "", "Alias for --filter")
	rootCmd.Flags().MarkHidden("fc")
	rootCmd.Flags().IntVar(&maxDepth, "depth", 1, "Recursive depth")
}

func main() {
        if err := rootCmd.Execute(); err != nil {
                os.Exit(1)
        }
}
