package auth_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"time"
)

var relevantCookies = map[string]bool{
	"FX_RM": true,
}

const CacheFile string = CacheDir + "/" + ".fantrax_cookie_cache.json"

func GetCookies() (string, error) {
	// First try environment variable
	if envCookies := os.Getenv("FANTRAX_COOKIES"); envCookies != "" {
		log.Debug("Found cookies from environment variable")
		return envCookies, nil
	}

	// Then try cache file
	cookies, err := getCookiesFromCache(CacheFile)
	if err == nil {
		log.Debug("Found cookies from cache")
		return convertCookiesToString(cookies)
	}

	// Finally fall back to browser
	log.Info("Fetching cookies with browser")
	cookies, err = GetCookiesWithBrowser(CacheFile)
	if err != nil {
		return "", err
	}

	return convertCookiesToString(cookies)
}

func convertCookiesToString(cookies []*network.Cookie) (string, error) {
	var cookieParts []string
	for _, c := range cookies {
		if _, ok := relevantCookies[c.Name]; !ok {
			continue
		}
		value := c.Value
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", c.Name, value))
	}

	if len(cookieParts) > 0 {
		return strings.Join(cookieParts, "; "), nil
	} else {
		return "", errors.New("no auth_client found")
	}
}

func getCookiesFromCache(cacheFile string) ([]*network.Cookie, error) {
	f, err := os.Open(cacheFile)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var cookies []*network.Cookie
	err = json.Unmarshal(data, &cookies)
	if err != nil {
		return nil, err
	}

	return cookies, nil
}

func GetCookiesWithBrowser(cacheFile string) ([]*network.Cookie, error) {
	// Get credentials from environment variables or command line
	username := os.Getenv("FANTRAX_USERNAME")
	password := os.Getenv("FANTRAX_PASSWORD")
	if username == "" || password == "" {
		return nil, errors.New("unable to fetch cookies from Fantrax." +
			"FANTRAX_USERNAME and FANTRAX_PASSWORD must be set as environment variables")
	}

	// Create a new Chrome instance in headless mode
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("window-size", "1920,1080"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create a new browser context with logging
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Set a timeout for the entire operation
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	fmt.Println("Navigating to Fantrax login page...")
	// Get the cookies after login
	var chromeCookies []*network.Cookie

	// Navigate to the login page and perform login
	err := chromedp.Run(ctx,
		// Navigate to the login page
		chromedp.Navigate("https://www.fantrax.com/login"),

		// Wait for the form to load completely
		chromedp.WaitVisible(`input[formcontrolname="email"]`),
		chromedp.WaitVisible(`input[formcontrolname="password"]`),

		// Enter email/username
		chromedp.Focus(`input[formcontrolname="email"]`),
		chromedp.SendKeys(`input[formcontrolname="email"]`, username),

		// Enter password
		chromedp.Focus(`input[formcontrolname="password"]`),
		chromedp.SendKeys(`input[formcontrolname="password"]`, password),

		// Wait a bit for validation to complete
		chromedp.Sleep(100*time.Millisecond),

		// Click the login button - using a more specific selector
		chromedp.Click(`button[type="submit"]`),

		// Wait for login to complete (could be navigation or a specific element)
		chromedp.Sleep(5*time.Second),
	)
	if err != nil {
		log.Fatalf("Login error: %v", err)
	}

	fmt.Println("Login successful. Getting auth_client...")

	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Get cookies from Chrome
		chromeCookies, err = storage.GetCookies().Do(ctx)
		if err != nil {
			return err
		}

		return nil
	}))

	// Write our cookies to cache
	f, err := os.Create(cacheFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cookieBytes, err := json.Marshal(chromeCookies)
	if err != nil {
		return nil, err
	}

	_, err = f.Write(cookieBytes)
	if err != nil {
		return nil, err
	}

	return chromeCookies, nil
}
