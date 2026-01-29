// Package cache provides caching for version information
package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/cboone/tmux-package-status/internal/parser"
)

// CacheEntry represents a cached result
type CacheEntry struct {
	Versions  []*parser.VersionInfo `json:"versions"`
	Timestamp int64                 `json:"timestamp"`
	Directory string                `json:"directory"`
}

// Cache provides caching functionality
type Cache struct {
	ttl      int
	cacheDir string
}

// New creates a new Cache with the given TTL in seconds
func New(ttl int) *Cache {
	cacheDir := getCacheDir()
	return &Cache{
		ttl:      ttl,
		cacheDir: cacheDir,
	}
}

// Get retrieves cached versions for a directory, returns nil if not found or expired
func (c *Cache) Get(dir string) []*parser.VersionInfo {
	if c.ttl <= 0 {
		return nil
	}

	cacheFile := c.getCacheFile(dir)
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	// Check if expired
	age := time.Now().Unix() - entry.Timestamp
	if age > int64(c.ttl) {
		return nil
	}

	// Check if directory matches
	if entry.Directory != dir {
		return nil
	}

	return entry.Versions
}

// Set stores versions in cache
func (c *Cache) Set(dir string, versions []*parser.VersionInfo) error {
	if c.ttl <= 0 {
		return nil
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return err
	}

	entry := CacheEntry{
		Versions:  versions,
		Timestamp: time.Now().Unix(),
		Directory: dir,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	cacheFile := c.getCacheFile(dir)
	return os.WriteFile(cacheFile, data, 0644)
}

// Clear removes the cache for a directory
func (c *Cache) Clear(dir string) error {
	cacheFile := c.getCacheFile(dir)
	return os.Remove(cacheFile)
}

// ClearAll removes all cached data
func (c *Cache) ClearAll() error {
	return os.RemoveAll(c.cacheDir)
}

// getCacheFile returns the cache file path for a directory
func (c *Cache) getCacheFile(dir string) string {
	// Create a hash of the directory path
	// MD5 is used here only for generating short, deterministic filenames, not for security
	hash := md5.Sum([]byte(dir)) // #nosec G401 -- MD5 used for filename hashing, not cryptography
	hashStr := hex.EncodeToString(hash[:])
	return filepath.Join(c.cacheDir, hashStr+".json")
}

// getCacheDir returns the cache directory path
func getCacheDir() string {
	// Try XDG_CACHE_HOME first
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, "tmux-package-status")
	}

	// Fall back to ~/.cache
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/tmux-package-status"
	}
	return filepath.Join(home, ".cache", "tmux-package-status")
}
