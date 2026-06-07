package engine

import (
	"sync"
	fhttp "github.com/bogdanfinn/fhttp"
)

type ScanResult struct {
	URL           string
	StatusCode    int
	Location      string
	Message       string
	ContentLength int64
	Cookies       []*fhttp.Cookie
}

type HeaderLine struct {
	Key   string
	Value string
}

var (
	Paused    bool
	PauseMu   sync.Mutex
	PauseChan = make(chan struct{})
)
