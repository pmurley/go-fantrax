package auth_client

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
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
}

// NewClient creates a new instance of the auth_client
func NewClient(leagueId string, useCache bool) *Client {
	return &Client{
		Client:   http.Client{},
		LeagueID: leagueId,
		UseCache: useCache,
	}
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
