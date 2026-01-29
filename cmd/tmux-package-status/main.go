// tmux-package-status displays package version information for tmux status bar
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cboone/tmux-package-status/internal/cache"
	"github.com/cboone/tmux-package-status/internal/config"
	"github.com/cboone/tmux-package-status/internal/detector"
	"github.com/cboone/tmux-package-status/internal/formatter"
	"github.com/cboone/tmux-package-status/internal/parser"
)

// Version information (set at build time)
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	// Command line flags
	var (
		dir          string
		format       string
		managers     string
		maxItems     int
		cacheTTL     int
		noCache      bool
		clearCache   bool
		showVersion  bool
		showHelp     bool
		listManagers bool

		// Style flags
		iconColor      string
		versionColor   string
		separatorColor string
		bracketColor   string
		showIcon       bool
		showBrackets   bool
		separator      string
		prefix         string
		suffix         string
		versionPrefix  string
	)

	flag.StringVar(&dir, "dir", "", "Directory to scan (default: current directory)")
	flag.StringVar(&dir, "d", "", "Directory to scan (shorthand)")

	flag.StringVar(&format, "format", "", "Output format: tmux, plain, json (default: tmux)")
	flag.StringVar(&format, "f", "", "Output format (shorthand)")

	flag.StringVar(&managers, "managers", "", "Comma-separated list of package managers to show")
	flag.StringVar(&managers, "m", "", "Package managers (shorthand)")

	flag.IntVar(&maxItems, "max", 0, "Maximum items to show (0 = use default)")
	flag.IntVar(&cacheTTL, "cache-ttl", 0, "Cache TTL in seconds (0 = use default)")
	flag.BoolVar(&noCache, "no-cache", false, "Disable caching")
	flag.BoolVar(&clearCache, "clear-cache", false, "Clear cache and exit")

	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version (shorthand)")

	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showHelp, "h", false, "Show help (shorthand)")

	flag.BoolVar(&listManagers, "list-managers", false, "List supported package managers")

	// Style flags
	flag.StringVar(&iconColor, "icon-color", "", "Icon color (tmux format)")
	flag.StringVar(&versionColor, "version-color", "", "Version color (tmux format)")
	flag.StringVar(&separatorColor, "separator-color", "", "Separator color (tmux format)")
	flag.StringVar(&bracketColor, "bracket-color", "", "Bracket color (tmux format)")
	flag.BoolVar(&showIcon, "icons", true, "Show icons")
	flag.BoolVar(&showBrackets, "brackets", false, "Show brackets around versions")
	flag.StringVar(&separator, "separator", "", "Separator between items")
	flag.StringVar(&prefix, "prefix", "", "Prefix for output")
	flag.StringVar(&suffix, "suffix", "", "Suffix for output")
	flag.StringVar(&versionPrefix, "version-prefix", "", "Prefix for version numbers")

	flag.Usage = printUsage
	flag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("tmux-package-status %s\n", Version)
		fmt.Printf("  commit: %s\n", Commit)
		fmt.Printf("  built:  %s\n", BuildDate)
		os.Exit(0)
	}

	if listManagers {
		printManagers()
		os.Exit(0)
	}

	// Load configuration from environment, then override with flags
	cfg := config.FromEnv()

	// Override with command line flags
	if dir != "" {
		cfg.Directory = dir
	}
	if format != "" {
		cfg.OutputFormat = format
	}
	if managers != "" {
		cfg.EnabledManagers = splitAndTrim(managers, ",")
	}
	if maxItems > 0 {
		cfg.MaxItems = maxItems
	}
	if cacheTTL > 0 {
		cfg.CacheTTL = cacheTTL
	}
	if noCache {
		cfg.CacheTTL = 0
	}

	// Style overrides
	if iconColor != "" {
		cfg.Style.IconColor = iconColor
	}
	if versionColor != "" {
		cfg.Style.VersionColor = versionColor
	}
	if separatorColor != "" {
		cfg.Style.SeparatorColor = separatorColor
	}
	if bracketColor != "" {
		cfg.Style.BracketColor = bracketColor
	}
	// Note: can't distinguish "not set" from "false" with bool flags easily
	// so we only have the positive case for these
	if flag.Lookup("icons").Value.String() == "false" {
		cfg.Style.ShowIcon = false
	}
	if flag.Lookup("brackets").Value.String() == "true" {
		cfg.Style.ShowBrackets = true
	}
	if separator != "" {
		cfg.Style.Separator = separator
	}
	if prefix != "" {
		cfg.Style.Prefix = prefix
	}
	if suffix != "" {
		cfg.Style.Suffix = suffix
	}
	if versionPrefix != "" {
		cfg.Style.VersionPrefix = versionPrefix
	}

	// Resolve directory
	if cfg.Directory == "" {
		var err error
		cfg.Directory, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not get current directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		var err error
		cfg.Directory, err = filepath.Abs(cfg.Directory)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not resolve directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Handle cache clear
	if clearCache {
		c := cache.New(cfg.CacheTTL)
		if err := c.ClearAll(); err != nil {
			fmt.Fprintf(os.Stderr, "error: could not clear cache: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Cache cleared")
		os.Exit(0)
	}

	// Run the main logic
	output, err := run(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}

// run executes the main logic and returns the formatted output
func run(cfg *config.Config) (string, error) {
	// Check cache first
	c := cache.New(cfg.CacheTTL)
	if versions := c.Get(cfg.Directory); versions != nil {
		// Filter by enabled managers
		versions = filterVersions(versions, cfg)
		f := formatter.New(cfg)
		return f.Format(versions), nil
	}

	// Detect package files
	d := detector.New(cfg.Directory)
	files, err := d.Detect()
	if err != nil {
		return "", fmt.Errorf("detection failed: %w", err)
	}

	if len(files) == 0 {
		return "", nil
	}

	// Parse versions
	p := parser.New()
	var versions []*parser.VersionInfo
	for _, file := range files {
		if !cfg.IsManagerEnabled(file.Type) {
			continue
		}

		v, err := p.Parse(file)
		if err != nil {
			// Log error but continue
			continue
		}
		if v != nil {
			versions = append(versions, v)
		}
	}

	// Cache the results (all versions, not filtered)
	c.Set(cfg.Directory, versions)

	// Format output
	f := formatter.New(cfg)
	return f.Format(versions), nil
}

// filterVersions filters versions by enabled managers
func filterVersions(versions []*parser.VersionInfo, cfg *config.Config) []*parser.VersionInfo {
	if len(cfg.EnabledManagers) == 0 {
		return versions
	}

	var filtered []*parser.VersionInfo
	for _, v := range versions {
		if cfg.IsManagerEnabled(v.Type) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func printUsage() {
	fmt.Println("tmux-package-status - Display package version information for tmux status bar")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  tmux-package-status [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -d, --dir <path>         Directory to scan (default: current directory)")
	fmt.Println("  -f, --format <format>    Output format: tmux, plain, json (default: tmux)")
	fmt.Println("  -m, --managers <list>    Comma-separated list of package managers to show")
	fmt.Println("      --max <n>            Maximum items to show (default: 5)")
	fmt.Println("      --cache-ttl <sec>    Cache TTL in seconds (default: 30)")
	fmt.Println("      --no-cache           Disable caching")
	fmt.Println("      --clear-cache        Clear cache and exit")
	fmt.Println()
	fmt.Println("  Style Options:")
	fmt.Println("      --icon-color <color>      Icon color (tmux format, e.g., colour39)")
	fmt.Println("      --version-color <color>   Version color")
	fmt.Println("      --separator-color <color> Separator color")
	fmt.Println("      --bracket-color <color>   Bracket color")
	fmt.Println("      --icons                   Show icons (default: true)")
	fmt.Println("      --brackets                Show brackets around versions")
	fmt.Println("      --separator <sep>         Separator between items (default: space)")
	fmt.Println("      --prefix <prefix>         Prefix for output")
	fmt.Println("      --suffix <suffix>         Suffix for output")
	fmt.Println("      --version-prefix <pre>    Prefix for version numbers (e.g., 'v')")
	fmt.Println()
	fmt.Println("  Other:")
	fmt.Println("  -v, --version            Show version information")
	fmt.Println("  -h, --help               Show this help message")
	fmt.Println("      --list-managers      List supported package managers")
	fmt.Println()
	fmt.Println("ENVIRONMENT VARIABLES:")
	fmt.Println("  TMUX_PKG_DIR             Directory to scan")
	fmt.Println("  TMUX_PKG_FORMAT          Output format")
	fmt.Println("  TMUX_PKG_MANAGERS        Comma-separated list of managers")
	fmt.Println("  TMUX_PKG_MAX_ITEMS       Maximum items to show")
	fmt.Println("  TMUX_PKG_CACHE_TTL       Cache TTL in seconds")
	fmt.Println("  TMUX_PKG_ICON_COLOR      Icon color")
	fmt.Println("  TMUX_PKG_VERSION_COLOR   Version color")
	fmt.Println("  TMUX_PKG_SHOW_ICON       Show icons (true/false)")
	fmt.Println("  TMUX_PKG_SHOW_BRACKETS   Show brackets (true/false)")
	fmt.Println("  TMUX_PKG_SEPARATOR       Separator between items")
	fmt.Println("  TMUX_PKG_PREFIX          Output prefix")
	fmt.Println("  TMUX_PKG_SUFFIX          Output suffix")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Basic usage in tmux.conf:")
	fmt.Println("  set -g status-right '#{E:tmux_package_status}'")
	fmt.Println()
	fmt.Println("  # Show only Node.js and Go versions:")
	fmt.Println("  tmux-package-status -m node,go")
	fmt.Println()
	fmt.Println("  # Plain text output:")
	fmt.Println("  tmux-package-status -f plain")
	fmt.Println()
	fmt.Println("  # JSON output:")
	fmt.Println("  tmux-package-status -f json")
}

func printManagers() {
	fmt.Println("Supported package managers:")
	fmt.Println()
	managers := []struct {
		name  string
		files string
		icon  string
	}{
		{"node", "package.json, .nvmrc, .node-version", "⬢"},
		{"go", "go.mod", "🐹"},
		{"rust", "Cargo.toml, rust-toolchain.toml", "🦀"},
		{"python", "pyproject.toml, requirements.txt, .python-version, Pipfile", "🐍"},
		{"ruby", "Gemfile, .ruby-version", "💎"},
		{"php", "composer.json", "🐘"},
		{"java", "pom.xml, build.gradle", "☕"},
		{"dotnet", "*.csproj, *.fsproj, global.json", "🔷"},
		{"elixir", "mix.exs", "💧"},
		{"deno", "deno.json, deno.jsonc", "🦕"},
		{"bun", "bun.lockb, bunfig.toml", "🥟"},
		{"zig", "build.zig, build.zig.zon", "⚡"},
		{"swift", "Package.swift", "🐦"},
		{"kotlin", "build.gradle.kts", "K"},
		{"scala", "build.sbt", "S"},
		{"haskell", "stack.yaml, *.cabal", "λ"},
		{"clojure", "project.clj, deps.edn", "λ"},
		{"lua", "*.rockspec", "🌙"},
		{"perl", "cpanfile, Makefile.PL", "🐪"},
	}

	for _, m := range managers {
		fmt.Printf("  %s %s\n", m.icon, m.name)
		fmt.Printf("    Files: %s\n", m.files)
	}
}
