// Package detector finds package manager files in a directory
package detector

import (
	"os"
	"path/filepath"
)

// PackageFile represents a detected package manager file
type PackageFile struct {
	// Type identifies the package manager (node, go, rust, python, ruby, etc.)
	Type string

	// Path is the full path to the package file
	Path string

	// Filename is just the filename (e.g., "package.json")
	Filename string

	// Priority determines display order (lower = higher priority)
	Priority int
}

// KnownFiles maps filenames to their package manager type and priority
var KnownFiles = map[string]struct {
	Type     string
	Priority int
}{
	// Multi-tool version managers (highest priority - these override individual files)
	// These are handled specially in Detect() to extract multiple versions
	".tool-versions":   {"tool-versions", 1},
	".mise.toml":       {"mise", 2},
	".mise.local.toml": {"mise", 2},
	".rtx.toml":        {"mise", 2},

	// Node.js ecosystem
	"package.json":  {"node", 10},
	".nvmrc":        {"node", 11},
	".node-version": {"node", 12},

	// Go
	"go.mod":      {"go", 20},
	".go-version": {"go", 21},

	// Rust
	"Cargo.toml": {"rust", 30},

	// Python
	"pyproject.toml":   {"python", 40},
	"requirements.txt": {"python", 41},
	"setup.py":         {"python", 42},
	"Pipfile":          {"python", 43},
	".python-version":  {"python", 44},
	"poetry.lock":      {"python", 45},

	// Ruby
	"Gemfile":        {"ruby", 50},
	".ruby-version":  {"ruby", 51},
	".rvmrc":         {"ruby", 52},
	"gems.rb":        {"ruby", 53},

	// PHP
	"composer.json": {"php", 60},

	// Java/JVM
	"pom.xml":      {"java", 70},
	"build.gradle": {"java", 71},
	"build.sbt":    {"scala", 72},

	// .NET
	"*.csproj":         {"dotnet", 80},
	"*.fsproj":         {"dotnet", 81},
	"packages.config":  {"dotnet", 82},
	"global.json":      {"dotnet", 83},

	// Elixir
	"mix.exs": {"elixir", 90},

	// Lua
	"*.rockspec": {"lua", 100},

	// Perl
	"cpanfile":   {"perl", 110},
	"Makefile.PL": {"perl", 111},

	// Swift
	"Package.swift": {"swift", 120},

	// Kotlin
	"build.gradle.kts": {"kotlin", 130},

	// Haskell
	"stack.yaml":    {"haskell", 140},
	"*.cabal":       {"haskell", 141},
	"cabal.project": {"haskell", 142},

	// Clojure
	"project.clj": {"clojure", 150},
	"deps.edn":    {"clojure", 151},

	// Zig
	"build.zig":     {"zig", 160},
	"build.zig.zon": {"zig", 161},

	// Deno
	"deno.json":  {"deno", 170},
	"deno.jsonc": {"deno", 171},

	// Bun
	"bun.lockb":   {"bun", 180},
	"bunfig.toml": {"bun", 181},
}

// Detector finds package files in directories
type Detector struct {
	rootDir string
}

// New creates a new Detector for the given directory
func New(dir string) *Detector {
	if dir == "" {
		dir, _ = os.Getwd()
	}
	return &Detector{rootDir: dir}
}

// Detect finds all package manager files in the directory
func (d *Detector) Detect() ([]PackageFile, error) {
	var results []PackageFile
	seen := make(map[string]bool) // track seen types to avoid duplicates

	// First check exact matches in root directory
	for filename, info := range KnownFiles {
		// Skip glob patterns in first pass
		if containsGlob(filename) {
			continue
		}

		fullPath := filepath.Join(d.rootDir, filename)
		if _, err := os.Stat(fullPath); err == nil {
			if !seen[info.Type] {
				results = append(results, PackageFile{
					Type:     info.Type,
					Path:     fullPath,
					Filename: filename,
					Priority: info.Priority,
				})
				seen[info.Type] = true
			}
		}
	}

	// Check glob patterns
	for pattern, info := range KnownFiles {
		if !containsGlob(pattern) {
			continue
		}

		if seen[info.Type] {
			continue
		}

		matches, err := filepath.Glob(filepath.Join(d.rootDir, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			if !seen[info.Type] {
				results = append(results, PackageFile{
					Type:     info.Type,
					Path:     match,
					Filename: filepath.Base(match),
					Priority: info.Priority,
				})
				seen[info.Type] = true
				break // only need one match per type
			}
		}
	}

	// Sort by priority
	sortByPriority(results)

	return results, nil
}

// DetectTypes returns just the types of package managers found
func (d *Detector) DetectTypes() ([]string, error) {
	files, err := d.Detect()
	if err != nil {
		return nil, err
	}

	types := make([]string, len(files))
	for i, f := range files {
		types[i] = f.Type
	}
	return types, nil
}

func containsGlob(pattern string) bool {
	for _, c := range pattern {
		if c == '*' || c == '?' || c == '[' {
			return true
		}
	}
	return false
}

func sortByPriority(files []PackageFile) {
	// Simple bubble sort - list is small
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[j].Priority < files[i].Priority {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}
