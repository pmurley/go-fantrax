package fantrax

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Cache defines the interface for a caching system
type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, data []byte) error
	GenerateKey(endpoint string, params map[string]string) string
}

// FileCache implements a file-based cache
type FileCache struct {
	CacheDir string
	TTL      time.Duration
}

// NewFileCache creates a new file-based cache
func NewFileCache(cacheDir string, ttl time.Duration) (*FileCache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return &FileCache{
		CacheDir: cacheDir,
		TTL:      ttl,
	}, nil
}

// GenerateKey creates a unique cache key
func (fc *FileCache) GenerateKey(endpoint string, params map[string]string) string {
	data := endpoint

	// Sort keys for consistency
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		data += fmt.Sprintf(":%s=%s", k, params[k])
	}

	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Get retrieves data from the cache if it exists and is not expired
func (fc *FileCache) Get(key string) ([]byte, bool) {
	cacheFile := filepath.Join(fc.CacheDir, key+".json")

	// Check if file exists
	fileInfo, err := os.Stat(cacheFile)
	if err != nil {
		return nil, false // Cache miss
	}

	// Check if cache is expired
	if fc.TTL > 0 && time.Since(fileInfo.ModTime()) > fc.TTL {
		return nil, false // Cache expired
	}

	// Read cache file
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false // Failed to read
	}

	return data, true
}

// Set stores data in the cache
func (fc *FileCache) Set(key string, data []byte) error {
	cacheFile := filepath.Join(fc.CacheDir, key+".json")
	return os.WriteFile(cacheFile, data, 0644)
}
