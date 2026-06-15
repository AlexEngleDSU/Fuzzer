package main

import (
    "fmt"
    "os"
    "github.com/AlexEngleDSU/Fuzzer/pkg/cli"
    "github.com/AlexEngleDSU/Fuzzer/pkg/ui"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: fuzzer [ui|cli] [options]")
        os.Exit(1)
    }

    switch os.Args[1] {
    case "ui":
        ui.Run()
    case "cli":
        // Pass the remaining arguments to the CLI logic
        cli.Run(os.Args[2:])
    default:
        fmt.Printf("Unknown command: %s\n", os.Args[1])
        os.Exit(1)
    }
}
