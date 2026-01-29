package config

import (
	"os"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.OutputFormat != "tmux" {
		t.Errorf("expected output format 'tmux', got '%s'", cfg.OutputFormat)
	}

	if cfg.MaxItems != 5 {
		t.Errorf("expected max items 5, got %d", cfg.MaxItems)
	}

	if cfg.CacheTTL != 30 {
		t.Errorf("expected cache TTL 30, got %d", cfg.CacheTTL)
	}

	if !cfg.Style.ShowIcon {
		t.Error("expected ShowIcon to be true")
	}

	if cfg.Style.ShowBrackets {
		t.Error("expected ShowBrackets to be false")
	}
}

func TestDefaultIcons(t *testing.T) {
	icons := DefaultIcons()

	testCases := []struct {
		key      string
		expected string
	}{
		{"node", "⬢"},
		{"go", "🐹"},
		{"rust", "🦀"},
		{"python", "🐍"},
		{"ruby", "💎"},
	}

	for _, tc := range testCases {
		if icons[tc.key] != tc.expected {
			t.Errorf("expected icon for %s to be '%s', got '%s'", tc.key, tc.expected, icons[tc.key])
		}
	}
}

func TestFromEnv(t *testing.T) {
	// Save and restore environment
	envVars := []string{
		"TMUX_PKG_DIR",
		"TMUX_PKG_MANAGERS",
		"TMUX_PKG_FORMAT",
		"TMUX_PKG_MAX_ITEMS",
		"TMUX_PKG_CACHE_TTL",
		"TMUX_PKG_ICON_COLOR",
		"TMUX_PKG_SHOW_ICON",
		"TMUX_PKG_SHOW_BRACKETS",
		"TMUX_PKG_SEPARATOR",
	}
	saved := make(map[string]string)
	for _, k := range envVars {
		saved[k] = os.Getenv(k)
	}
	defer func() {
		for k, v := range saved {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Set test values
	os.Setenv("TMUX_PKG_DIR", "/test/dir")
	os.Setenv("TMUX_PKG_MANAGERS", "node, go, rust")
	os.Setenv("TMUX_PKG_FORMAT", "json")
	os.Setenv("TMUX_PKG_MAX_ITEMS", "10")
	os.Setenv("TMUX_PKG_CACHE_TTL", "60")
	os.Setenv("TMUX_PKG_ICON_COLOR", "colour200")
	os.Setenv("TMUX_PKG_SHOW_ICON", "false")
	os.Setenv("TMUX_PKG_SHOW_BRACKETS", "true")
	os.Setenv("TMUX_PKG_SEPARATOR", " | ")

	cfg := FromEnv()

	if cfg.Directory != "/test/dir" {
		t.Errorf("expected directory '/test/dir', got '%s'", cfg.Directory)
	}

	if len(cfg.EnabledManagers) != 3 {
		t.Errorf("expected 3 enabled managers, got %d", len(cfg.EnabledManagers))
	}

	if cfg.OutputFormat != "json" {
		t.Errorf("expected format 'json', got '%s'", cfg.OutputFormat)
	}

	if cfg.MaxItems != 10 {
		t.Errorf("expected max items 10, got %d", cfg.MaxItems)
	}

	if cfg.CacheTTL != 60 {
		t.Errorf("expected cache TTL 60, got %d", cfg.CacheTTL)
	}

	if cfg.Style.IconColor != "colour200" {
		t.Errorf("expected icon color 'colour200', got '%s'", cfg.Style.IconColor)
	}

	if cfg.Style.ShowIcon {
		t.Error("expected ShowIcon to be false")
	}

	if !cfg.Style.ShowBrackets {
		t.Error("expected ShowBrackets to be true")
	}

	if cfg.Style.Separator != " | " {
		t.Errorf("expected separator ' | ', got '%s'", cfg.Style.Separator)
	}
}

func TestIsManagerEnabled(t *testing.T) {
	cfg := Default()

	// Empty list means all enabled
	if !cfg.IsManagerEnabled("node") {
		t.Error("expected node to be enabled with empty list")
	}
	if !cfg.IsManagerEnabled("go") {
		t.Error("expected go to be enabled with empty list")
	}

	// With specific list
	cfg.EnabledManagers = []string{"node", "go"}

	if !cfg.IsManagerEnabled("node") {
		t.Error("expected node to be enabled")
	}
	if !cfg.IsManagerEnabled("go") {
		t.Error("expected go to be enabled")
	}
	if cfg.IsManagerEnabled("rust") {
		t.Error("expected rust to be disabled")
	}

	// Case insensitive
	if !cfg.IsManagerEnabled("NODE") {
		t.Error("expected NODE (uppercase) to match")
	}
	if !cfg.IsManagerEnabled("Go") {
		t.Error("expected Go (mixed case) to match")
	}
}
