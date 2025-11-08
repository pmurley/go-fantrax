package auth_client

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pmurley/go-fantrax"
	"github.com/pmurley/go-fantrax/models"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
)

const CacheDir string = "./.fantrax-cache"

type FantraxRequest struct {
	Msgs []FantraxMessage `json:"msgs"`
}

type FantraxMessage struct {
	Method string      `json:"method"`
	Data   interface{} `json:"data"`
}

type Client struct {
	http.Client
	LeagueID string
	UseCache bool
	UserInfo *models.UserInfo
}

// NewClient creates a new instance of the auth_client and fetches user info
func NewClient(leagueId string, useCache bool) (*Client, error) {
	client := &Client{
		Client:   http.Client{},
		LeagueID: leagueId,
		UseCache: useCache,
	}

	// Fetch user info including timezone data
	err := client.Login()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info during client initialization: %w", err)
	}

	return client, nil
}

// Do sends an HTTP request and returns an HTTP response
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var cacheKey string
	var newBody io.ReadCloser
	var err error
	if c.UseCache {
		cacheKey, newBody, err = hashReadCloser(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = newBody
		log.Info("cache key: ", cacheKey)

		info, err := os.Stat(path.Join(CacheDir, cacheKey))

		if err == nil && info.Size() > 0 {
			cachedResponse, err := os.Open(path.Join(CacheDir, cacheKey))
			if err != nil {
				return nil, err
			}

			// Read the file content
			cachedData, err := io.ReadAll(cachedResponse)
			if err != nil {
				cachedResponse.Close()
				return nil, err
			}
			cachedResponse.Close()

			// Create a new reader from the data
			response := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(cachedData)),
			}
			log.Info("cache hit")
			return response, nil
		}
		log.Info("cache miss")
	}

	cookiesString, err := GetCookies()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", cookiesString)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if c.UseCache {
		// Read the entire response body
		respData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		resp.Body.Close()

		// Write to cache file
		err = os.MkdirAll(CacheDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache directory: %w", err)
		}

		err = os.WriteFile(path.Join(CacheDir, cacheKey), respData, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to write cache file: %w", err)
		}

		// Create a new response body for the consumer
		resp.Body = io.NopCloser(bytes.NewBuffer(respData))
	}

	return resp, nil
}

func hashReadCloser(rc io.ReadCloser) (string, io.ReadCloser, error) {
	// Read all bytes from the reader
	body, err := io.ReadAll(rc)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read content: %w", err)
	}

	// Create a new reader with the same content for the caller
	newReader := io.NopCloser(bytes.NewBuffer(body))

	// Calculate MD5 hash
	hash := md5.Sum(body)
	hashStr := hex.EncodeToString(hash[:])

	return hashStr, newReader, nil
}

// LoginResponse represents the structure of the login API response
type LoginResponse struct {
	Responses []struct {
		Data struct {
			UserInfo models.UserInfo `json:"userInfo"`
		} `json:"data"`
	} `json:"responses"`
}

// Login calls the login endpoint and stores user info including timezone data
func (c *Client) Login() error {
	// Build the request
	fullRequest := map[string]interface{}{
		"msgs": []FantraxMessage{
			{
				Method: "login",
				Data:   map[string]interface{}{},
			},
		},
		"uiv":    3,
		"refUrl": fmt.Sprintf("https://www.fantrax.com/newui/fantasy/miscellaneous.go?leagueId=%s", c.LeagueID),
		"dt":     0,
		"at":     0,
		"av":     "0.0",
		"tz":     "UTC",
		"v":      "167.0.1",
	}

	jsonStr, err := json.Marshal(fullRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://www.fantrax.com/fxpa/req?leagueId="+c.LeagueID, bytes.NewBuffer(jsonStr))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response body: %w", err)
	}

	var loginResponse LoginResponse
	err = json.Unmarshal(body, &loginResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	if len(loginResponse.Responses) == 0 {
		return fmt.Errorf("no responses in login response")
	}

	// Store the user info in the client
	c.UserInfo = &loginResponse.Responses[0].Data.UserInfo

	return nil
}

// GetCurrentPeriod fetches the current scoring period from Fantrax
// This uses the public API to get the current period number
func (c *Client) GetCurrentPeriod() (int, error) {
	// Use the public fantrax client to get rosters which includes the current period
	publicClient, err := fantrax.NewClient(c.LeagueID, false)
	if err != nil {
		return 0, fmt.Errorf("failed to create public client: %w", err)
	}

	rosters, err := publicClient.GetTeamRosters()
	if err != nil {
		return 0, fmt.Errorf("failed to get team rosters: %w", err)
	}

	return rosters.Period, nil
}
