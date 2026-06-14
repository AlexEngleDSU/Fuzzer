package browser

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/playwright-community/playwright-go"
	tls_client "github.com/bogdanfinn/tls-client"
)

// Cookie represents the structure for injection.
type Cookie struct {
	Name  string
	Value string
}

func EnsureEnvironment() {
	if BrowserExists() {
		return
	}
	fmt.Println("First run detected: Installing browser dependencies...")
	os.Setenv("PLAYWRIGHT_BROWSERS_PATH", GetBrowserPath())
	if err := playwright.Install(); err != nil {
		fmt.Printf("Error installing browsers: %v\n", err)
		return
	}
	// Optimized: Remove non-essential engines to save space
	os.RemoveAll(filepath.Join(GetBrowserPath(), "firefox"))
	os.RemoveAll(filepath.Join(GetBrowserPath(), "webkit"))
	fmt.Println("Browser dependencies installed and optimized.")
}

func BrowserExists() bool {
	info, err := os.Stat(GetBrowserPath())
	return err == nil && info.IsDir()
}

func GetBrowserPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	// Simplified map-based path resolution
	paths := map[string]string{
		"windows": "AppData/Local/ms-playwright",
		"darwin":  "Library/Caches/ms-playwright",
	}
	
	relPath, ok := paths[runtime.GOOS]
	if !ok {
		relPath = ".cache/ms-playwright"
	}
	return filepath.Join(home, relPath)
}

func InjectCookies(jar tls_client.CookieJar, targetURL string, cookies []Cookie) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return
	}

	var fhttpCookies []*fhttp.Cookie
	for _, c := range cookies {
		fhttpCookies = append(fhttpCookies, &fhttp.Cookie{
			Name:   c.Name,
			Value:  c.Value,
			Domain: u.Host,
			Path:   "/",
		})
	}
	jar.SetCookies(u, fhttpCookies)
}
