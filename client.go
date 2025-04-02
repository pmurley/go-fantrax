package fantrax

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io"
	"net/http"
	"time"
)

const CachePath string = "./fantrax-cache"

// Client represents a Fantrax API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client

	Cache        *FileCache
	CacheEnabled bool
}

// NewClient creates a new Fantrax API client
func NewClient(cacheEnabled bool) (*Client, error) {
	client := &Client{
		BaseURL:      "https://www.fantrax.com/fxea",
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
		CacheEnabled: cacheEnabled,
	}

	// Initialize cache if enabled
	if cacheEnabled {
		cache, err := NewFileCache(CachePath, 24*time.Hour)
		if err != nil {
			return nil, err
		}
		client.Cache = cache
	}
	return client, nil
}

// fetchWithCache is a helper method that handles caching logic
func (c *Client) fetchWithCache(endpoint string, params map[string]string, result interface{}) error {
	// If caching is disabled, make a direct request
	if !c.CacheEnabled || c.Cache == nil {
		return c.makeRequest(endpoint, params, result)
	}

	// Generate cache key
	cacheKey := c.Cache.GenerateKey(endpoint, params)

	// Try to get from cache
	if cachedData, found := c.Cache.Get(cacheKey); found {
		// Unmarshal cached data
		fmt.Printf("Cache hit: %s\n", cacheKey)
		return json.Unmarshal(cachedData, result)
	}

	fmt.Printf("Cache miss: %s\n", cacheKey)
	// Cache miss - make the request
	var responseData []byte
	err := c.makeRequestRaw(endpoint, params, &responseData)
	if err != nil {
		return err
	}

	// Store in cache
	if err := c.Cache.Set(cacheKey, responseData); err != nil {
		// Log but don't fail the request
		fmt.Printf("Failed to cache response: %v\n", err)
	}

	// Unmarshal the response
	return json.Unmarshal(responseData, result)
}

// makeRequestRaw makes an API request and returns the raw response body
func (c *Client) makeRequestRaw(endpoint string, params map[string]string, responseData *[]byte) error {
	// Build URL with query parameters
	url := c.BaseURL + endpoint

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	// Make the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	// Read the entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	*responseData = body
	return nil
}

// makeRequest makes an API request and unmarshals the response into result
func (c *Client) makeRequest(endpoint string, params map[string]string, result interface{}) error {
	var responseData []byte
	if err := c.makeRequestRaw(endpoint, params, &responseData); err != nil {
		return err
	}

	if err := json.Unmarshal(responseData, result); err != nil {
		spew.Dump(responseData)
		return err
	}
	return nil
}
