package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cboone/tmux-package-status/internal/detector"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Error("expected non-nil Parser")
	}
}

func TestParseNode_Nvmrc(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .nvmrc file
	nvmrcPath := filepath.Join(tmpDir, ".nvmrc")
	if err := os.WriteFile(nvmrcPath, []byte("v18.17.0\n"), 0644); err != nil {
		t.Fatalf("failed to create .nvmrc: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "node",
		Path:     nvmrcPath,
		Filename: ".nvmrc",
		Priority: 11,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v == nil {
		t.Fatal("expected version info, got nil")
	}

	if v.Version != "18.17.0" {
		t.Errorf("expected version '18.17.0', got '%s'", v.Version)
	}

	if v.Source != "file:.nvmrc" {
		t.Errorf("expected source 'file:.nvmrc', got '%s'", v.Source)
	}
}

func TestParseNode_NodeVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .node-version file
	path := filepath.Join(tmpDir, ".node-version")
	if err := os.WriteFile(path, []byte("20.10.0"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "node",
		Path:     path,
		Filename: ".node-version",
		Priority: 12,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != "20.10.0" {
		t.Errorf("expected version '20.10.0', got '%s'", v.Version)
	}
}

func TestParseNode_PackageJsonEngines(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json with engines
	pkgJSON := `{
		"name": "test",
		"engines": {
			"node": ">=18.0.0"
		}
	}`
	path := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(path, []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "node",
		Path:     path,
		Filename: "package.json",
		Priority: 10,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v == nil {
		t.Fatal("expected version info")
	}

	if v.Version != ">=18.0.0" {
		t.Errorf("expected version '>=18.0.0', got '%s'", v.Version)
	}
}

func TestParseGo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goMod := `module example.com/test

go 1.21

require (
	github.com/something v1.0.0
)`
	path := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(path, []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "go",
		Path:     path,
		Filename: "go.mod",
		Priority: 20,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v == nil {
		t.Fatal("expected version info")
	}

	if v.Version != "1.21" {
		t.Errorf("expected version '1.21', got '%s'", v.Version)
	}

	if v.Source != "go.mod" {
		t.Errorf("expected source 'go.mod', got '%s'", v.Source)
	}
}

func TestParseGo_WithPatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goMod := `module test
go 1.21.5`
	path := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(path, []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "go",
		Path:     path,
		Filename: "go.mod",
		Priority: 20,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != "1.21.5" {
		t.Errorf("expected version '1.21.5', got '%s'", v.Version)
	}
}

func TestParseRust_CargoToml(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cargoToml := `[package]
name = "test"
version = "0.1.0"
rust-version = "1.70"`
	path := filepath.Join(tmpDir, "Cargo.toml")
	if err := os.WriteFile(path, []byte(cargoToml), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "rust",
		Path:     path,
		Filename: "Cargo.toml",
		Priority: 30,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v == nil {
		t.Fatal("expected version info")
	}

	if v.Version != "1.70" {
		t.Errorf("expected version '1.70', got '%s'", v.Version)
	}
}

func TestParseRust_Toolchain(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rust-toolchain.toml
	toolchain := `[toolchain]
channel = "1.75.0"`
	toolchainPath := filepath.Join(tmpDir, "rust-toolchain.toml")
	if err := os.WriteFile(toolchainPath, []byte(toolchain), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Create Cargo.toml
	cargoPath := filepath.Join(tmpDir, "Cargo.toml")
	if err := os.WriteFile(cargoPath, []byte("[package]\nname = \"test\""), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "rust",
		Path:     cargoPath,
		Filename: "Cargo.toml",
		Priority: 30,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Should get version from rust-toolchain.toml
	if v.Version != "1.75.0" {
		t.Errorf("expected version '1.75.0', got '%s'", v.Version)
	}

	if v.Source != "rust-toolchain.toml" {
		t.Errorf("expected source 'rust-toolchain.toml', got '%s'", v.Source)
	}
}

func TestParsePython_Version(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, ".python-version")
	if err := os.WriteFile(path, []byte("3.11.5"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "python",
		Path:     path,
		Filename: ".python-version",
		Priority: 44,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != "3.11.5" {
		t.Errorf("expected version '3.11.5', got '%s'", v.Version)
	}
}

func TestParsePython_Pyproject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	pyproject := `[project]
name = "test"
requires-python = ">=3.10"`
	path := filepath.Join(tmpDir, "pyproject.toml")
	if err := os.WriteFile(path, []byte(pyproject), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "python",
		Path:     path,
		Filename: "pyproject.toml",
		Priority: 40,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != ">=3.10" {
		t.Errorf("expected version '>=3.10', got '%s'", v.Version)
	}
}

func TestParseRuby_Version(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, ".ruby-version")
	if err := os.WriteFile(path, []byte("3.2.2"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "ruby",
		Path:     path,
		Filename: ".ruby-version",
		Priority: 51,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != "3.2.2" {
		t.Errorf("expected version '3.2.2', got '%s'", v.Version)
	}
}

func TestParseRuby_VersionWithPrefix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, ".ruby-version")
	if err := os.WriteFile(path, []byte("ruby-3.2.2"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "ruby",
		Path:     path,
		Filename: ".ruby-version",
		Priority: 51,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Should strip ruby- prefix
	if v.Version != "3.2.2" {
		t.Errorf("expected version '3.2.2', got '%s'", v.Version)
	}
}

func TestParseRuby_Gemfile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gemfile := `source "https://rubygems.org"
ruby "3.1.0"

gem "rails", "~> 7.0"`
	path := filepath.Join(tmpDir, "Gemfile")
	if err := os.WriteFile(path, []byte(gemfile), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "ruby",
		Path:     path,
		Filename: "Gemfile",
		Priority: 50,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != "3.1.0" {
		t.Errorf("expected version '3.1.0', got '%s'", v.Version)
	}
}

func TestParsePHP(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	composer := `{
		"require": {
			"php": ">=8.1"
		}
	}`
	path := filepath.Join(tmpDir, "composer.json")
	if err := os.WriteFile(path, []byte(composer), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "php",
		Path:     path,
		Filename: "composer.json",
		Priority: 60,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != ">=8.1" {
		t.Errorf("expected version '>=8.1', got '%s'", v.Version)
	}
}

func TestParseElixir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mixExs := `defmodule MyApp.MixProject do
  use Mix.Project

  def project do
    [
      app: :my_app,
      version: "0.1.0",
      elixir: "~> 1.15",
      start_permanent: Mix.env() == :prod,
      deps: deps()
    ]
  end
end`
	path := filepath.Join(tmpDir, "mix.exs")
	if err := os.WriteFile(path, []byte(mixExs), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "elixir",
		Path:     path,
		Filename: "mix.exs",
		Priority: 90,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != "~> 1.15" {
		t.Errorf("expected version '~> 1.15', got '%s'", v.Version)
	}
}

func TestParseSwift(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	pkgSwift := `// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "MyPackage"
)`
	path := filepath.Join(tmpDir, "Package.swift")
	if err := os.WriteFile(path, []byte(pkgSwift), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "swift",
		Path:     path,
		Filename: "Package.swift",
		Priority: 120,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != "5.9" {
		t.Errorf("expected version '5.9', got '%s'", v.Version)
	}
}

func TestParseScala(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	buildSbt := `name := "my-project"
version := "0.1.0"
scalaVersion := "3.3.1"`
	path := filepath.Join(tmpDir, "build.sbt")
	if err := os.WriteFile(path, []byte(buildSbt), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "scala",
		Path:     path,
		Filename: "build.sbt",
		Priority: 72,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != "3.3.1" {
		t.Errorf("expected version '3.3.1', got '%s'", v.Version)
	}
}

func TestParseHaskell(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stackYaml := `resolver: lts-21.25
packages:
- .`
	path := filepath.Join(tmpDir, "stack.yaml")
	if err := os.WriteFile(path, []byte(stackYaml), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "haskell",
		Path:     path,
		Filename: "stack.yaml",
		Priority: 140,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != "lts-21.25" {
		t.Errorf("expected version 'lts-21.25', got '%s'", v.Version)
	}
}

func TestParse_UnknownType(t *testing.T) {
	p := New()
	pf := detector.PackageFile{
		Type:     "unknown",
		Path:     "/nonexistent",
		Filename: "unknown.file",
		Priority: 999,
	}

	v, err := p.Parse(pf)
	// Should not error, just return nil if no version found
	if err != nil && v != nil {
		t.Errorf("unexpected result for unknown type")
	}
}

func TestParse_FileNotFound(t *testing.T) {
	p := New()
	pf := detector.PackageFile{
		Type:     "node",
		Path:     "/nonexistent/.nvmrc",
		Filename: ".nvmrc",
		Priority: 11,
	}

	_, err := p.Parse(pf)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseToolVersions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	toolVersions := `# This is a comment
nodejs 18.17.0
python 3.11.5
ruby 3.2.2
golang 1.21.0
`
	path := filepath.Join(tmpDir, ".tool-versions")
	if err := os.WriteFile(path, []byte(toolVersions), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "tool-versions",
		Path:     path,
		Filename: ".tool-versions",
		Priority: 1,
	}

	versions, err := p.ParseMulti(pf)
	if err != nil {
		t.Fatalf("ParseMulti() failed: %v", err)
	}

	if len(versions) != 4 {
		t.Errorf("expected 4 versions, got %d", len(versions))
	}

	// Check that tools were parsed correctly
	typeVersions := make(map[string]string)
	for _, v := range versions {
		typeVersions[v.Type] = v.Version
	}

	if typeVersions["node"] != "18.17.0" {
		t.Errorf("expected node version '18.17.0', got '%s'", typeVersions["node"])
	}
	if typeVersions["python"] != "3.11.5" {
		t.Errorf("expected python version '3.11.5', got '%s'", typeVersions["python"])
	}
	if typeVersions["ruby"] != "3.2.2" {
		t.Errorf("expected ruby version '3.2.2', got '%s'", typeVersions["ruby"])
	}
	if typeVersions["go"] != "1.21.0" {
		t.Errorf("expected go version '1.21.0', got '%s'", typeVersions["go"])
	}
}

func TestParseToolVersions_MultipleVersions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// asdf supports multiple versions per tool
	toolVersions := `nodejs 18.17.0 20.10.0
python 3.11.5
`
	path := filepath.Join(tmpDir, ".tool-versions")
	if err := os.WriteFile(path, []byte(toolVersions), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "tool-versions",
		Path:     path,
		Filename: ".tool-versions",
		Priority: 1,
	}

	versions, err := p.ParseMulti(pf)
	if err != nil {
		t.Fatalf("ParseMulti() failed: %v", err)
	}

	// Should take first version for node
	typeVersions := make(map[string]string)
	for _, v := range versions {
		typeVersions[v.Type] = v.Version
	}

	if typeVersions["node"] != "18.17.0" {
		t.Errorf("expected node version '18.17.0' (first listed), got '%s'", typeVersions["node"])
	}
}

func TestParseMise(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	miseToml := `[tools]
nodejs = "18.17.0"
python = "3.11.5"
golang = "1.21"
`
	path := filepath.Join(tmpDir, ".mise.toml")
	if err := os.WriteFile(path, []byte(miseToml), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "mise",
		Path:     path,
		Filename: ".mise.toml",
		Priority: 2,
	}

	versions, err := p.ParseMulti(pf)
	if err != nil {
		t.Fatalf("ParseMulti() failed: %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(versions))
	}

	typeVersions := make(map[string]string)
	for _, v := range versions {
		typeVersions[v.Type] = v.Version
	}

	if typeVersions["node"] != "18.17.0" {
		t.Errorf("expected node version '18.17.0', got '%s'", typeVersions["node"])
	}
	if typeVersions["python"] != "3.11.5" {
		t.Errorf("expected python version '3.11.5', got '%s'", typeVersions["python"])
	}
	if typeVersions["go"] != "1.21" {
		t.Errorf("expected go version '1.21', got '%s'", typeVersions["go"])
	}
}

func TestParseMise_ArrayFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// mise also supports array format
	miseToml := `[tools]
nodejs = ["18.17.0", "20.10.0"]
python = "3.11.5"
`
	path := filepath.Join(tmpDir, ".mise.toml")
	if err := os.WriteFile(path, []byte(miseToml), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "mise",
		Path:     path,
		Filename: ".mise.toml",
		Priority: 2,
	}

	versions, err := p.ParseMulti(pf)
	if err != nil {
		t.Fatalf("ParseMulti() failed: %v", err)
	}

	typeVersions := make(map[string]string)
	for _, v := range versions {
		typeVersions[v.Type] = v.Version
	}

	// Should take first version from array
	if typeVersions["node"] != "18.17.0" {
		t.Errorf("expected node version '18.17.0' (first in array), got '%s'", typeVersions["node"])
	}
}

func TestParseGo_GoVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, ".go-version")
	if err := os.WriteFile(path, []byte("1.21.5"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "go",
		Path:     path,
		Filename: ".go-version",
		Priority: 21,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if v.Version != "1.21.5" {
		t.Errorf("expected version '1.21.5', got '%s'", v.Version)
	}
}

func TestParseGo_GoVersionWithPrefix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-parse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Some tools prefix with "go"
	path := filepath.Join(tmpDir, ".go-version")
	if err := os.WriteFile(path, []byte("go1.21.5"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	p := New()
	pf := detector.PackageFile{
		Type:     "go",
		Path:     path,
		Filename: ".go-version",
		Priority: 21,
	}

	v, err := p.Parse(pf)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Should strip "go" prefix
	if v.Version != "1.21.5" {
		t.Errorf("expected version '1.21.5', got '%s'", v.Version)
	}
}

func TestMapToolName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"nodejs", "node"},
		{"node", "node"},
		{"golang", "go"},
		{"go", "go"},
		{"python", "python"},
		{"python3", "python"},
		{"ruby", "ruby"},
		{"rust", "rust"},
		{"java", "java"},
		{"openjdk", "java"},
		{"unknown", ""},
		{"NODEJS", "node"}, // case insensitive
	}

	for _, tc := range tests {
		result := mapToolName(tc.input)
		if result != tc.expected {
			t.Errorf("mapToolName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
