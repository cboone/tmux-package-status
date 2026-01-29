package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cboone/tmux-package-status/internal/parser"
)

func TestNew(t *testing.T) {
	c := New(30)
	if c == nil {
		t.Error("expected non-nil Cache")
	}
	if c.ttl != 30 {
		t.Errorf("expected TTL 30, got %d", c.ttl)
	}
}

func TestCache_SetAndGet(t *testing.T) {
	// Use a temp directory for cache
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &Cache{
		ttl:      60,
		cacheDir: tmpDir,
	}

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
		{Type: "go", Version: "1.21", Source: "go.mod", Priority: 20},
	}

	testDir := "/test/project"

	// Set cache
	if err := c.Set(testDir, versions); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Get cache
	result := c.Get(testDir)
	if result == nil {
		t.Fatal("expected cached versions, got nil")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 cached versions, got %d", len(result))
	}

	if result[0].Type != "node" || result[0].Version != "18.17.0" {
		t.Errorf("unexpected first cached version: %+v", result[0])
	}
}

func TestCache_GetNonexistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &Cache{
		ttl:      60,
		cacheDir: tmpDir,
	}

	result := c.Get("/nonexistent/dir")
	if result != nil {
		t.Errorf("expected nil for nonexistent cache, got %v", result)
	}
}

func TestCache_Expired(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Very short TTL
	c := &Cache{
		ttl:      1,
		cacheDir: tmpDir,
	}

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}

	testDir := "/test/project"
	if err := c.Set(testDir, versions); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Should work immediately
	result := c.Get(testDir)
	if result == nil {
		t.Error("expected cached versions immediately after set")
	}

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Should be expired now
	result = c.Get(testDir)
	if result != nil {
		t.Error("expected nil for expired cache")
	}
}

func TestCache_DisabledTTL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// TTL of 0 disables caching
	c := &Cache{
		ttl:      0,
		cacheDir: tmpDir,
	}

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}

	testDir := "/test/project"

	// Set should not error but should not actually cache
	if err := c.Set(testDir, versions); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Get should return nil
	result := c.Get(testDir)
	if result != nil {
		t.Error("expected nil when TTL is 0")
	}
}

func TestCache_Clear(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &Cache{
		ttl:      60,
		cacheDir: tmpDir,
	}

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}

	testDir := "/test/project"
	c.Set(testDir, versions)

	// Verify cache exists
	if c.Get(testDir) == nil {
		t.Fatal("expected cache to exist before clear")
	}

	// Clear
	if err := c.Clear(testDir); err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	// Should be gone
	if c.Get(testDir) != nil {
		t.Error("expected cache to be cleared")
	}
}

func TestCache_ClearAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &Cache{
		ttl:      60,
		cacheDir: tmpDir,
	}

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}

	// Set multiple caches
	c.Set("/test/project1", versions)
	c.Set("/test/project2", versions)

	// Clear all
	if err := c.ClearAll(); err != nil {
		t.Fatalf("ClearAll() failed: %v", err)
	}

	// Both should be gone
	if c.Get("/test/project1") != nil {
		t.Error("expected cache 1 to be cleared")
	}
	if c.Get("/test/project2") != nil {
		t.Error("expected cache 2 to be cleared")
	}

	// Cache directory should not exist
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Error("expected cache directory to be removed")
	}
}

func TestCache_DifferentDirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c := &Cache{
		ttl:      60,
		cacheDir: tmpDir,
	}

	versions1 := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}
	versions2 := []*parser.VersionInfo{
		{Type: "go", Version: "1.21", Source: "go.mod", Priority: 20},
	}

	c.Set("/project1", versions1)
	c.Set("/project2", versions2)

	// Should get different results for different directories
	result1 := c.Get("/project1")
	result2 := c.Get("/project2")

	if result1[0].Type != "node" {
		t.Errorf("expected node for project1, got %s", result1[0].Type)
	}
	if result2[0].Type != "go" {
		t.Errorf("expected go for project2, got %s", result2[0].Type)
	}
}

func TestGetCacheDir(t *testing.T) {
	// Test with XDG_CACHE_HOME set
	oldXDG := os.Getenv("XDG_CACHE_HOME")
	defer os.Setenv("XDG_CACHE_HOME", oldXDG)

	os.Setenv("XDG_CACHE_HOME", "/custom/cache")
	dir := getCacheDir()
	expected := "/custom/cache/tmux-package-status"
	if dir != expected {
		t.Errorf("expected '%s', got '%s'", expected, dir)
	}

	// Test without XDG_CACHE_HOME
	os.Unsetenv("XDG_CACHE_HOME")
	dir = getCacheDir()
	home, _ := os.UserHomeDir()
	expected = filepath.Join(home, ".cache", "tmux-package-status")
	if dir != expected {
		t.Errorf("expected '%s', got '%s'", expected, dir)
	}
}

func TestCache_GetCacheFile(t *testing.T) {
	c := &Cache{
		ttl:      60,
		cacheDir: "/test/cache",
	}

	file1 := c.getCacheFile("/project1")
	file2 := c.getCacheFile("/project2")

	// Different directories should produce different cache files
	if file1 == file2 {
		t.Error("expected different cache files for different directories")
	}

	// Same directory should produce same cache file
	file1b := c.getCacheFile("/project1")
	if file1 != file1b {
		t.Error("expected same cache file for same directory")
	}
}
