// Package parser extracts version information from package files
package parser

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cboone/tmux-package-status/internal/detector"
)

// VersionInfo holds version information for a package manager
type VersionInfo struct {
	// Type is the package manager type (node, go, rust, etc.)
	Type string

	// Version is the detected version
	Version string

	// Source indicates where the version came from (file, runtime, etc.)
	Source string

	// Priority for display ordering
	Priority int
}

// Parser extracts version information from package files
type Parser struct{}

// New creates a new Parser
func New() *Parser {
	return &Parser{}
}

// Parse extracts version information from a package file
func (p *Parser) Parse(pf detector.PackageFile) (*VersionInfo, error) {
	var version string
	var source string
	var err error

	switch pf.Type {
	case "tool-versions", "mise":
		// These are multi-tool files, use ParseMulti instead
		return nil, nil
	case "node":
		version, source, err = p.parseNode(pf)
	case "go":
		version, source, err = p.parseGo(pf)
	case "rust":
		version, source, err = p.parseRust(pf)
	case "python":
		version, source, err = p.parsePython(pf)
	case "ruby":
		version, source, err = p.parseRuby(pf)
	case "php":
		version, source, err = p.parsePHP(pf)
	case "java":
		version, source, err = p.parseJava(pf)
	case "dotnet":
		version, source, err = p.parseDotNet(pf)
	case "elixir":
		version, source, err = p.parseElixir(pf)
	case "deno":
		version, source, err = p.parseDeno(pf)
	case "bun":
		version, source, err = p.parseBun(pf)
	case "zig":
		version, source, err = p.parseZig(pf)
	case "swift":
		version, source, err = p.parseSwift(pf)
	case "haskell":
		version, source, err = p.parseHaskell(pf)
	case "scala":
		version, source, err = p.parseScala(pf)
	case "kotlin":
		version, source, err = p.parseKotlin(pf)
	case "clojure":
		version, source, err = p.parseClojure(pf)
	case "lua":
		version, source, err = p.parseLua(pf)
	case "perl":
		version, source, err = p.parsePerl(pf)
	default:
		// Try to get runtime version
		version, source, err = p.getRuntimeVersion(pf.Type)
	}

	if err != nil {
		return nil, err
	}

	if version == "" {
		return nil, nil
	}

	return &VersionInfo{
		Type:     pf.Type,
		Version:  version,
		Source:   source,
		Priority: pf.Priority,
	}, nil
}

// ParseMulti extracts multiple version infos from multi-tool files like .tool-versions
func (p *Parser) ParseMulti(pf detector.PackageFile) ([]*VersionInfo, error) {
	switch pf.Type {
	case "tool-versions":
		return p.parseToolVersions(pf)
	case "mise":
		return p.parseMise(pf)
	default:
		// Fall back to single parse
		v, err := p.Parse(pf)
		if err != nil || v == nil {
			return nil, err
		}
		return []*VersionInfo{v}, nil
	}
}

// parseToolVersions parses .tool-versions files (asdf/mise format)
// Format: tool_name version [version2 ...]
func (p *Parser) parseToolVersions(pf detector.PackageFile) ([]*VersionInfo, error) {
	content, err := os.ReadFile(pf.Path)
	if err != nil {
		return nil, err
	}

	var versions []*VersionInfo
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		toolName := parts[0]
		version := parts[1]

		// Map asdf/mise tool names to our internal types
		internalType := mapToolName(toolName)
		if internalType == "" {
			continue
		}

		versions = append(versions, &VersionInfo{
			Type:     internalType,
			Version:  version,
			Source:   ".tool-versions",
			Priority: pf.Priority,
		})
	}

	return versions, nil
}

// parseMise parses .mise.toml files
// Format: [tools]\nnodejs = "18.17.0"\npython = "3.11"
func (p *Parser) parseMise(pf detector.PackageFile) ([]*VersionInfo, error) {
	content, err := os.ReadFile(pf.Path)
	if err != nil {
		return nil, err
	}

	var versions []*VersionInfo
	contentStr := string(content)

	// Simple TOML parsing for [tools] section
	// Look for patterns like: tool_name = "version" or tool_name = ["version1", "version2"]
	inToolsSection := false
	lines := strings.Split(contentStr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for section headers
		if strings.HasPrefix(line, "[") {
			inToolsSection = strings.HasPrefix(line, "[tools]")
			continue
		}

		if !inToolsSection {
			continue
		}

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first =
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		toolName := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])

		var version string

		// Handle array format: ["version1", "version2"]
		if strings.HasPrefix(valueStr, "[") {
			// Extract first element from array
			re := regexp.MustCompile(`\[\s*"([^"]+)"`)
			if matches := re.FindStringSubmatch(valueStr); len(matches) > 1 {
				version = matches[1]
			} else {
				// Try without quotes
				re = regexp.MustCompile(`\[\s*([^\s,\]]+)`)
				if matches := re.FindStringSubmatch(valueStr); len(matches) > 1 {
					version = matches[1]
				}
			}
		} else {
			// Handle simple format: "version" or version
			version = strings.Trim(valueStr, `"'`)
		}

		if version == "" {
			continue
		}

		internalType := mapToolName(toolName)
		if internalType == "" {
			continue
		}

		versions = append(versions, &VersionInfo{
			Type:     internalType,
			Version:  version,
			Source:   pf.Filename,
			Priority: pf.Priority,
		})
	}

	return versions, nil
}

// mapToolName converts asdf/mise tool names to internal type names
func mapToolName(toolName string) string {
	// Map of asdf/mise plugin names to our internal types
	mapping := map[string]string{
		// Node.js variants
		"nodejs":  "node",
		"node":    "node",
		"npm":     "node",

		// Go variants
		"golang": "go",
		"go":     "go",

		// Python variants
		"python":  "python",
		"python3": "python",

		// Ruby
		"ruby": "ruby",

		// Rust
		"rust":  "rust",
		"cargo": "rust",

		// PHP
		"php": "php",

		// Java variants
		"java":    "java",
		"openjdk": "java",
		"adoptopenjdk": "java",
		"temurin": "java",

		// .NET
		"dotnet":      "dotnet",
		"dotnet-core": "dotnet",

		// Elixir/Erlang
		"elixir": "elixir",
		"erlang": "erlang",

		// Deno
		"deno": "deno",

		// Bun
		"bun": "bun",

		// Zig
		"zig": "zig",

		// Swift
		"swift": "swift",

		// Kotlin
		"kotlin": "kotlin",

		// Scala
		"scala": "scala",

		// Haskell
		"haskell": "haskell",
		"ghc":     "haskell",

		// Clojure
		"clojure": "clojure",

		// Lua
		"lua":      "lua",
		"luajit":   "lua",

		// Perl
		"perl": "perl",

		// Additional tools
		"terraform": "terraform",
		"kubectl":   "kubectl",
		"helm":      "helm",
	}

	if internal, ok := mapping[strings.ToLower(toolName)]; ok {
		return internal
	}
	return ""
}

// parseNode extracts Node.js version from package files
func (p *Parser) parseNode(pf detector.PackageFile) (string, string, error) {
	switch pf.Filename {
	case ".nvmrc", ".node-version":
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}
		version := strings.TrimSpace(string(content))
		version = strings.TrimPrefix(version, "v")
		return version, "file:" + pf.Filename, nil

	case "package.json":
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		var pkg struct {
			Engines struct {
				Node string `json:"node"`
			} `json:"engines"`
		}
		if err := json.Unmarshal(content, &pkg); err == nil && pkg.Engines.Node != "" {
			return pkg.Engines.Node, "engines", nil
		}

		// Fall back to runtime version
		return p.getCommandVersion("node", "--version", "v")
	}

	return p.getCommandVersion("node", "--version", "v")
}

// parseGo extracts Go version from go.mod or .go-version
func (p *Parser) parseGo(pf detector.PackageFile) (string, string, error) {
	switch pf.Filename {
	case ".go-version":
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}
		version := strings.TrimSpace(string(content))
		version = strings.TrimPrefix(version, "go")
		return version, ".go-version", nil

	case "go.mod":
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`(?m)^go\s+(\d+\.\d+(?:\.\d+)?)`)
		matches := re.FindStringSubmatch(string(content))
		if len(matches) > 1 {
			return matches[1], "go.mod", nil
		}
	}

	return p.getCommandVersion("go", "version", "go version go")
}

// parseRust extracts Rust version from Cargo.toml or toolchain
func (p *Parser) parseRust(pf detector.PackageFile) (string, string, error) {
	// Get the directory containing the package file
	dir := filepath.Dir(pf.Path)

	// Check for rust-toolchain.toml first
	toolchainPath := filepath.Join(dir, "rust-toolchain.toml")
	if content, err := os.ReadFile(toolchainPath); err == nil {
		re := regexp.MustCompile(`channel\s*=\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "rust-toolchain.toml", nil
		}
	}

	// Check for rust-toolchain file
	toolchainPath = filepath.Join(dir, "rust-toolchain")
	if content, err := os.ReadFile(toolchainPath); err == nil {
		version := strings.TrimSpace(string(content))
		if version != "" {
			return version, "rust-toolchain", nil
		}
	}

	// Check Cargo.toml for rust-version
	if pf.Filename == "Cargo.toml" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`(?m)^rust-version\s*=\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "Cargo.toml", nil
		}
	}

	return p.getCommandVersion("rustc", "--version", "rustc ")
}

// parsePython extracts Python version
func (p *Parser) parsePython(pf detector.PackageFile) (string, string, error) {
	switch pf.Filename {
	case ".python-version":
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}
		version := strings.TrimSpace(string(content))
		return version, ".python-version", nil

	case "pyproject.toml":
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		// Check requires-python
		re := regexp.MustCompile(`(?m)requires-python\s*=\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "pyproject.toml", nil
		}

		// Check python version in tool.poetry
		re = regexp.MustCompile(`(?m)python\s*=\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "pyproject.toml", nil
		}
	}

	return p.getCommandVersion("python3", "--version", "Python ")
}

// parseRuby extracts Ruby version
func (p *Parser) parseRuby(pf detector.PackageFile) (string, string, error) {
	switch pf.Filename {
	case ".ruby-version":
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}
		version := strings.TrimSpace(string(content))
		version = strings.TrimPrefix(version, "ruby-")
		return version, ".ruby-version", nil

	case "Gemfile":
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`(?m)ruby\s+['"]([^'"]+)['"]`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "Gemfile", nil
		}
	}

	return p.getCommandVersion("ruby", "--version", "ruby ")
}

// parsePHP extracts PHP version
func (p *Parser) parsePHP(pf detector.PackageFile) (string, string, error) {
	if pf.Filename == "composer.json" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		var pkg struct {
			Require struct {
				PHP string `json:"php"`
			} `json:"require"`
		}
		if err := json.Unmarshal(content, &pkg); err == nil && pkg.Require.PHP != "" {
			return pkg.Require.PHP, "composer.json", nil
		}
	}

	return p.getCommandVersion("php", "--version", "PHP ")
}

// parseJava extracts Java version
func (p *Parser) parseJava(pf detector.PackageFile) (string, string, error) {
	if pf.Filename == "pom.xml" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		// Check for maven.compiler.source or java.version
		re := regexp.MustCompile(`<(?:maven\.compiler\.source|java\.version)>([^<]+)</`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "pom.xml", nil
		}
	}

	if pf.Filename == "build.gradle" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`(?m)sourceCompatibility\s*=\s*['"]?([^\s'"]+)`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "build.gradle", nil
		}
	}

	return p.getCommandVersion("java", "-version", "version \"")
}

// parseDotNet extracts .NET version
func (p *Parser) parseDotNet(pf detector.PackageFile) (string, string, error) {
	if pf.Filename == "global.json" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		var pkg struct {
			SDK struct {
				Version string `json:"version"`
			} `json:"sdk"`
		}
		if err := json.Unmarshal(content, &pkg); err == nil && pkg.SDK.Version != "" {
			return pkg.SDK.Version, "global.json", nil
		}
	}

	if strings.HasSuffix(pf.Filename, ".csproj") || strings.HasSuffix(pf.Filename, ".fsproj") {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`<TargetFramework>net(\d+\.\d+)</TargetFramework>`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], pf.Filename, nil
		}
	}

	return p.getCommandVersion("dotnet", "--version", "")
}

// parseElixir extracts Elixir version from mix.exs
func (p *Parser) parseElixir(pf detector.PackageFile) (string, string, error) {
	if pf.Filename == "mix.exs" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`elixir:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "mix.exs", nil
		}
	}

	return p.getCommandVersion("elixir", "--version", "Elixir ")
}

// parseDeno extracts Deno version from deno.json
func (p *Parser) parseDeno(pf detector.PackageFile) (string, string, error) {
	return p.getCommandVersion("deno", "--version", "deno ")
}

// parseBun extracts Bun version
func (p *Parser) parseBun(pf detector.PackageFile) (string, string, error) {
	return p.getCommandVersion("bun", "--version", "")
}

// parseZig extracts Zig version
func (p *Parser) parseZig(pf detector.PackageFile) (string, string, error) {
	if pf.Filename == "build.zig.zon" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`\.minimum_zig_version\s*=\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "build.zig.zon", nil
		}
	}

	return p.getCommandVersion("zig", "version", "")
}

// parseSwift extracts Swift version
func (p *Parser) parseSwift(pf detector.PackageFile) (string, string, error) {
	if pf.Filename == "Package.swift" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`swift-tools-version:\s*(\d+\.\d+(?:\.\d+)?)`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "Package.swift", nil
		}
	}

	return p.getCommandVersion("swift", "--version", "Swift version ")
}

// parseHaskell extracts Haskell/GHC version
func (p *Parser) parseHaskell(pf detector.PackageFile) (string, string, error) {
	if pf.Filename == "stack.yaml" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`(?m)^resolver:\s*(\S+)`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "stack.yaml", nil
		}
	}

	return p.getCommandVersion("ghc", "--version", "version ")
}

// parseScala extracts Scala version
func (p *Parser) parseScala(pf detector.PackageFile) (string, string, error) {
	if pf.Filename == "build.sbt" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`scalaVersion\s*:=\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "build.sbt", nil
		}
	}

	return p.getCommandVersion("scala", "-version", "version ")
}

// parseKotlin extracts Kotlin version
func (p *Parser) parseKotlin(pf detector.PackageFile) (string, string, error) {
	if pf.Filename == "build.gradle.kts" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`kotlin\("jvm"\)\s*version\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "build.gradle.kts", nil
		}
	}

	return p.getCommandVersion("kotlin", "-version", "Kotlin version ")
}

// parseClojure extracts Clojure version
func (p *Parser) parseClojure(pf detector.PackageFile) (string, string, error) {
	if pf.Filename == "deps.edn" {
		content, err := os.ReadFile(pf.Path)
		if err != nil {
			return "", "", err
		}

		re := regexp.MustCompile(`org\.clojure/clojure\s*\{:mvn/version\s*"([^"]+)"\}`)
		if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], "deps.edn", nil
		}
	}

	return p.getCommandVersion("clj", "--version", "Clojure CLI version ")
}

// parseLua extracts Lua version
func (p *Parser) parseLua(pf detector.PackageFile) (string, string, error) {
	return p.getCommandVersion("lua", "-v", "Lua ")
}

// parsePerl extracts Perl version
func (p *Parser) parsePerl(pf detector.PackageFile) (string, string, error) {
	return p.getCommandVersion("perl", "--version", "v")
}

// getCommandVersion runs a command and extracts version
func (p *Parser) getCommandVersion(cmd string, args string, prefix string) (string, string, error) {
	out, err := exec.Command(cmd, args).CombinedOutput()
	if err != nil {
		return "", "", err
	}

	output := string(out)
	if prefix != "" {
		idx := strings.Index(output, prefix)
		if idx >= 0 {
			output = output[idx+len(prefix):]
		}
	}

	// Extract version number
	re := regexp.MustCompile(`^(\d+\.\d+(?:\.\d+)?(?:-[\w.]+)?)`)
	if matches := re.FindStringSubmatch(strings.TrimSpace(output)); len(matches) > 1 {
		return matches[1], "runtime", nil
	}

	// Just return first word if no version pattern found
	parts := strings.Fields(output)
	if len(parts) > 0 {
		return parts[0], "runtime", nil
	}

	return "", "", nil
}

// getRuntimeVersion gets version from runtime command
func (p *Parser) getRuntimeVersion(pkgType string) (string, string, error) {
	cmdMap := map[string]struct {
		cmd    string
		args   string
		prefix string
	}{
		"node":    {"node", "--version", "v"},
		"go":      {"go", "version", "go version go"},
		"rust":    {"rustc", "--version", "rustc "},
		"python":  {"python3", "--version", "Python "},
		"ruby":    {"ruby", "--version", "ruby "},
		"php":     {"php", "--version", "PHP "},
		"java":    {"java", "-version", "version \""},
		"dotnet":  {"dotnet", "--version", ""},
		"elixir":  {"elixir", "--version", "Elixir "},
		"deno":    {"deno", "--version", "deno "},
		"bun":     {"bun", "--version", ""},
		"zig":     {"zig", "version", ""},
		"swift":   {"swift", "--version", "Swift version "},
		"haskell": {"ghc", "--version", "version "},
		"scala":   {"scala", "-version", "version "},
		"kotlin":  {"kotlin", "-version", "Kotlin version "},
		"clojure": {"clj", "--version", "Clojure CLI version "},
		"lua":     {"lua", "-v", "Lua "},
		"perl":    {"perl", "--version", "v"},
	}

	if info, ok := cmdMap[pkgType]; ok {
		return p.getCommandVersion(info.cmd, info.args, info.prefix)
	}

	return "", "", nil
}
