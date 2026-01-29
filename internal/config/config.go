// Package config handles configuration for tmux-package-status
package config

import (
	"os"
	"strconv"
	"strings"
)

// StyleConfig holds styling configuration for output
type StyleConfig struct {
	// Colors (tmux format: colour0-colour255, or named colors)
	IconColor      string
	VersionColor   string
	SeparatorColor string
	BracketColor   string

	// Format options
	ShowIcon       bool
	ShowBrackets   bool
	Separator      string
	Prefix         string
	Suffix         string
	VersionPrefix  string

	// Icon overrides (per language/tool)
	Icons map[string]string
}

// Config holds the main configuration
type Config struct {
	// Directory to scan (defaults to current directory)
	Directory string

	// Which package managers to detect
	EnabledManagers []string

	// Styling
	Style StyleConfig

	// Output format: "tmux", "plain", "json"
	OutputFormat string

	// Max items to show (0 = unlimited)
	MaxItems int

	// Cache settings
	CacheTTL int // seconds, 0 = disabled
}

// DefaultIcons returns the default icons for each package manager
func DefaultIcons() map[string]string {
	return map[string]string{
		"node":      "⬢",
		"npm":       "📦",
		"go":        "🐹",
		"rust":      "🦀",
		"python":    "🐍",
		"ruby":      "💎",
		"php":       "🐘",
		"java":      "☕",
		"dotnet":    "🔷",
		"elixir":    "💧",
		"erlang":    "📡",
		"lua":       "🌙",
		"perl":      "🐪",
		"swift":     "🐦",
		"kotlin":    "K",
		"scala":     "S",
		"haskell":   "λ",
		"clojure":   "λ",
		"zig":       "⚡",
		"deno":      "🦕",
		"bun":       "🥟",
		"terraform": "🏗️",
		"kubectl":   "☸️",
		"helm":      "⎈",
	}
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Directory:       "",
		EnabledManagers: []string{}, // empty = all
		OutputFormat:    "tmux",
		MaxItems:        5,
		CacheTTL:        30,
		Style: StyleConfig{
			IconColor:      "colour39",  // bright blue
			VersionColor:   "colour250", // light gray
			SeparatorColor: "colour240", // dark gray
			BracketColor:   "colour240",
			ShowIcon:       true,
			ShowBrackets:   false,
			Separator:      " ",
			Prefix:         "",
			Suffix:         "",
			VersionPrefix:  "",
			Icons:          DefaultIcons(),
		},
	}
}

// FromEnv loads configuration from environment variables
// Environment variables use the prefix TMUX_PKG_
func FromEnv() *Config {
	cfg := Default()

	// Directory
	if dir := os.Getenv("TMUX_PKG_DIR"); dir != "" {
		cfg.Directory = dir
	}

	// Enabled managers (comma-separated)
	if managers := os.Getenv("TMUX_PKG_MANAGERS"); managers != "" {
		cfg.EnabledManagers = strings.Split(managers, ",")
		for i := range cfg.EnabledManagers {
			cfg.EnabledManagers[i] = strings.TrimSpace(cfg.EnabledManagers[i])
		}
	}

	// Output format
	if format := os.Getenv("TMUX_PKG_FORMAT"); format != "" {
		cfg.OutputFormat = format
	}

	// Max items
	if maxStr := os.Getenv("TMUX_PKG_MAX_ITEMS"); maxStr != "" {
		if max, err := strconv.Atoi(maxStr); err == nil {
			cfg.MaxItems = max
		}
	}

	// Cache TTL
	if ttlStr := os.Getenv("TMUX_PKG_CACHE_TTL"); ttlStr != "" {
		if ttl, err := strconv.Atoi(ttlStr); err == nil {
			cfg.CacheTTL = ttl
		}
	}

	// Style options
	if color := os.Getenv("TMUX_PKG_ICON_COLOR"); color != "" {
		cfg.Style.IconColor = color
	}
	if color := os.Getenv("TMUX_PKG_VERSION_COLOR"); color != "" {
		cfg.Style.VersionColor = color
	}
	if color := os.Getenv("TMUX_PKG_SEPARATOR_COLOR"); color != "" {
		cfg.Style.SeparatorColor = color
	}
	if color := os.Getenv("TMUX_PKG_BRACKET_COLOR"); color != "" {
		cfg.Style.BracketColor = color
	}

	if showIcon := os.Getenv("TMUX_PKG_SHOW_ICON"); showIcon != "" {
		cfg.Style.ShowIcon = showIcon == "true" || showIcon == "1"
	}
	if showBrackets := os.Getenv("TMUX_PKG_SHOW_BRACKETS"); showBrackets != "" {
		cfg.Style.ShowBrackets = showBrackets == "true" || showBrackets == "1"
	}

	if sep := os.Getenv("TMUX_PKG_SEPARATOR"); sep != "" {
		cfg.Style.Separator = sep
	}
	if prefix := os.Getenv("TMUX_PKG_PREFIX"); prefix != "" {
		cfg.Style.Prefix = prefix
	}
	if suffix := os.Getenv("TMUX_PKG_SUFFIX"); suffix != "" {
		cfg.Style.Suffix = suffix
	}
	if vprefix := os.Getenv("TMUX_PKG_VERSION_PREFIX"); vprefix != "" {
		cfg.Style.VersionPrefix = vprefix
	}

	// Icon overrides (TMUX_PKG_ICON_node, TMUX_PKG_ICON_go, etc.)
	for key, defaultIcon := range DefaultIcons() {
		envKey := "TMUX_PKG_ICON_" + strings.ToUpper(key)
		if icon := os.Getenv(envKey); icon != "" {
			cfg.Style.Icons[key] = icon
		} else {
			cfg.Style.Icons[key] = defaultIcon
		}
	}

	return cfg
}

// IsManagerEnabled checks if a package manager is enabled
func (c *Config) IsManagerEnabled(name string) bool {
	if len(c.EnabledManagers) == 0 {
		return true // all enabled if list is empty
	}
	for _, m := range c.EnabledManagers {
		if strings.EqualFold(m, name) {
			return true
		}
	}
	return false
}
