package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	d := New("/test/dir")
	if d.rootDir != "/test/dir" {
		t.Errorf("expected rootDir '/test/dir', got '%s'", d.rootDir)
	}

	d = New("")
	cwd, _ := os.Getwd()
	if d.rootDir != cwd {
		t.Errorf("expected rootDir to be current directory, got '%s'", d.rootDir)
	}
}

func TestDetect(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test package files
	testFiles := []string{
		"package.json",
		"go.mod",
		"Cargo.toml",
	}

	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", f, err)
		}
	}

	d := New(tmpDir)
	files, err := d.Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}

	// Check types
	types := make(map[string]bool)
	for _, f := range files {
		types[f.Type] = true
	}

	if !types["node"] {
		t.Error("expected to detect node")
	}
	if !types["go"] {
		t.Error("expected to detect go")
	}
	if !types["rust"] {
		t.Error("expected to detect rust")
	}
}

func TestDetect_Empty(t *testing.T) {
	// Create empty temp directory
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-test-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	d := New(tmpDir)
	files, err := d.Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestDetect_Priority(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-test-priority-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files with different priorities
	testFiles := []string{
		"Gemfile",       // ruby, priority 50
		"package.json",  // node, priority 10
		"Cargo.toml",    // rust, priority 30
	}

	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", f, err)
		}
	}

	d := New(tmpDir)
	files, err := d.Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	// Should be sorted by priority: node (10), rust (30), ruby (50)
	if files[0].Type != "node" {
		t.Errorf("expected first file to be node, got %s", files[0].Type)
	}
	if files[1].Type != "rust" {
		t.Errorf("expected second file to be rust, got %s", files[1].Type)
	}
	if files[2].Type != "ruby" {
		t.Errorf("expected third file to be ruby, got %s", files[2].Type)
	}
}

func TestDetect_NoDuplicateTypes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-test-nodup-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple node-related files
	testFiles := []string{
		"package.json",
		".nvmrc",
		".node-version",
	}

	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("18.0.0"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", f, err)
		}
	}

	d := New(tmpDir)
	files, err := d.Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	// Should only have one node entry (the one with lowest priority)
	nodeCount := 0
	for _, f := range files {
		if f.Type == "node" {
			nodeCount++
		}
	}

	if nodeCount != 1 {
		t.Errorf("expected 1 node entry, got %d", nodeCount)
	}
}

func TestDetectTypes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmux-pkg-test-types-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFiles := []string{"package.json", "go.mod"}
	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", f, err)
		}
	}

	d := New(tmpDir)
	types, err := d.DetectTypes()
	if err != nil {
		t.Fatalf("DetectTypes() failed: %v", err)
	}

	if len(types) != 2 {
		t.Errorf("expected 2 types, got %d", len(types))
	}
}

func TestContainsGlob(t *testing.T) {
	tests := []struct {
		pattern  string
		expected bool
	}{
		{"package.json", false},
		{"*.json", true},
		{"*.csproj", true},
		{"go.mod", false},
		{"src/**/*.ts", true},
		{"test[123].txt", true},
		{"file?.txt", true},
	}

	for _, tc := range tests {
		result := containsGlob(tc.pattern)
		if result != tc.expected {
			t.Errorf("containsGlob(%q) = %v, expected %v", tc.pattern, result, tc.expected)
		}
	}
}

func TestSortByPriority(t *testing.T) {
	files := []PackageFile{
		{Type: "ruby", Priority: 50},
		{Type: "node", Priority: 10},
		{Type: "go", Priority: 20},
		{Type: "rust", Priority: 30},
	}

	sortByPriority(files)

	expected := []string{"node", "go", "rust", "ruby"}
	for i, f := range files {
		if f.Type != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], f.Type)
		}
	}
}

func TestKnownFiles(t *testing.T) {
	// Verify essential files are in the map
	essentialFiles := []struct {
		filename string
		typ      string
	}{
		{"package.json", "node"},
		{"go.mod", "go"},
		{"Cargo.toml", "rust"},
		{"pyproject.toml", "python"},
		{"Gemfile", "ruby"},
		{"composer.json", "php"},
		{"pom.xml", "java"},
		{"mix.exs", "elixir"},
		{"deno.json", "deno"},
	}

	for _, ef := range essentialFiles {
		info, ok := KnownFiles[ef.filename]
		if !ok {
			t.Errorf("expected %s to be in KnownFiles", ef.filename)
			continue
		}
		if info.Type != ef.typ {
			t.Errorf("expected %s to have type %s, got %s", ef.filename, ef.typ, info.Type)
		}
	}
}
