package browser

import (
	"fmt"
	"github.com/playwright-community/playwright-go"
)

type WAFSession struct {
	Cookies []map[string]interface{}
	Headers map[string]string
}

func InitializeSession(targetURL string) (*WAFSession, error) {
	
	pw, err := playwright.Run(&playwright.RunOptions{ Verbose: true, })
	if err != nil {
		return nil, fmt.Errorf("failed to start playwright: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
	        "--no-sandbox",        
	        "--disable-setuid-sandbox",
	        "--disable-dev-shm-usage", 
    		},
	})
	if err != nil {
		pw.Stop() // Cleanup before exiting
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	// 3. Create Context and Page
	context, err := browser.NewContext()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to create context: %w", err)
	}

	page, err := context.NewPage()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// 4. Navigate
	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
            WaitUntil: playwright.WaitUntilStateNetworkidle,
        })
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to navigate to %s: %w", targetURL, err)
	}

	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	// 5. Extract Cookies
	rawCookies, err := context.Cookies()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to get cookies: %w", err)
	}

	var formattedCookies []map[string]interface{}
	for _, c := range rawCookies {
		formattedCookies = append(formattedCookies, map[string]interface{}{
			"name":   c.Name,
			"value":  c.Value,
			"domain": c.Domain,
		})
	}

	browser.Close()
	pw.Stop()

	return &WAFSession{Cookies: formattedCookies}, nil
}
