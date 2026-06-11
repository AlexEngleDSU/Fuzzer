package engine
import (
	"sync"
	fhttp "github.com/bogdanfinn/fhttp"
)
type WAFSession struct {
	Cookies []map[string]interface{}
	Headers map[string]string
}

type ScanState struct {
    Results []ScanResult
    Mu      sync.RWMutex
}

func (s *ScanState) Add(res ScanResult) {
    s.Mu.Lock()
    defer s.Mu.Unlock()
    s.Results = append(s.Results, res)
}

type ScanResult struct {
	URL               string
	StatusCode        int
	Location          string
	Message           string
	ContentLength     int64
	Cookies           []*fhttp.Cookie
	Depth		  int
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

