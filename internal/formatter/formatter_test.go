package formatter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cboone/tmux-package-status/internal/config"
	"github.com/cboone/tmux-package-status/internal/parser"
)

func TestNew(t *testing.T) {
	cfg := config.Default()
	f := New(cfg)
	if f == nil {
		t.Error("expected non-nil Formatter")
	}
}

func TestFormat_Empty(t *testing.T) {
	cfg := config.Default()
	f := New(cfg)

	result := f.Format(nil)
	if result != "" {
		t.Errorf("expected empty string for nil input, got '%s'", result)
	}

	result = f.Format([]*parser.VersionInfo{})
	if result != "" {
		t.Errorf("expected empty string for empty input, got '%s'", result)
	}
}

func TestFormat_Tmux(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "tmux"
	f := New(cfg)

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}

	result := f.Format(versions)

	// Should contain icon and version with color codes
	if !strings.Contains(result, "18.17.0") {
		t.Errorf("expected version in output, got '%s'", result)
	}
	if !strings.Contains(result, "#[fg=") {
		t.Errorf("expected tmux color codes in output, got '%s'", result)
	}
	if !strings.Contains(result, "⬢") {
		t.Errorf("expected node icon in output, got '%s'", result)
	}
}

func TestFormat_TmuxMultiple(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "tmux"
	f := New(cfg)

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
		{Type: "go", Version: "1.21", Source: "go.mod", Priority: 20},
	}

	result := f.Format(versions)

	if !strings.Contains(result, "18.17.0") {
		t.Errorf("expected node version in output, got '%s'", result)
	}
	if !strings.Contains(result, "1.21") {
		t.Errorf("expected go version in output, got '%s'", result)
	}
	if !strings.Contains(result, "⬢") {
		t.Errorf("expected node icon in output, got '%s'", result)
	}
	if !strings.Contains(result, "🐹") {
		t.Errorf("expected go icon in output, got '%s'", result)
	}
}

func TestFormat_Plain(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "plain"
	f := New(cfg)

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}

	result := f.Format(versions)

	// Should NOT contain tmux color codes
	if strings.Contains(result, "#[fg=") {
		t.Errorf("expected no tmux color codes in plain output, got '%s'", result)
	}
	if !strings.Contains(result, "18.17.0") {
		t.Errorf("expected version in output, got '%s'", result)
	}
	if !strings.Contains(result, "⬢") {
		t.Errorf("expected icon in output, got '%s'", result)
	}
}

func TestFormat_PlainNoIcons(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "plain"
	cfg.Style.ShowIcon = false
	f := New(cfg)

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}

	result := f.Format(versions)

	if strings.Contains(result, "⬢") {
		t.Errorf("expected no icon when ShowIcon=false, got '%s'", result)
	}
	if !strings.Contains(result, "18.17.0") {
		t.Errorf("expected version in output, got '%s'", result)
	}
}

func TestFormat_JSON(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "json"
	f := New(cfg)

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
		{Type: "go", Version: "1.21", Source: "go.mod", Priority: 20},
	}

	result := f.Format(versions)

	// Should be valid JSON
	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v, got '%s'", err, result)
	}

	if len(parsed) != 2 {
		t.Errorf("expected 2 items in JSON, got %d", len(parsed))
	}

	if parsed[0]["type"] != "node" {
		t.Errorf("expected first type 'node', got '%v'", parsed[0]["type"])
	}
	if parsed[0]["version"] != "18.17.0" {
		t.Errorf("expected first version '18.17.0', got '%v'", parsed[0]["version"])
	}
}

func TestFormat_MaxItems(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "plain"
	cfg.MaxItems = 2
	f := New(cfg)

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
		{Type: "go", Version: "1.21", Source: "go.mod", Priority: 20},
		{Type: "rust", Version: "1.70", Source: "Cargo.toml", Priority: 30},
		{Type: "python", Version: "3.11", Source: "file", Priority: 40},
	}

	result := f.Format(versions)

	// Should only contain first 2 items
	if !strings.Contains(result, "18.17.0") {
		t.Errorf("expected node version in output")
	}
	if !strings.Contains(result, "1.21") {
		t.Errorf("expected go version in output")
	}
	if strings.Contains(result, "1.70") {
		t.Errorf("expected rust version to be excluded due to MaxItems")
	}
	if strings.Contains(result, "3.11") {
		t.Errorf("expected python version to be excluded due to MaxItems")
	}
}

func TestFormat_WithBrackets(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "plain"
	cfg.Style.ShowBrackets = true
	cfg.Style.ShowIcon = false
	f := New(cfg)

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}

	result := f.Format(versions)

	if !strings.Contains(result, "[18.17.0]") {
		t.Errorf("expected bracketed version in output, got '%s'", result)
	}
}

func TestFormat_WithPrefix(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "plain"
	cfg.Style.Prefix = ">> "
	cfg.Style.Suffix = " <<"
	f := New(cfg)

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}

	result := f.Format(versions)

	if !strings.HasPrefix(result, ">> ") {
		t.Errorf("expected prefix '>> ', got '%s'", result)
	}
	if !strings.HasSuffix(result, " <<") {
		t.Errorf("expected suffix ' <<', got '%s'", result)
	}
}

func TestFormat_WithVersionPrefix(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "plain"
	cfg.Style.ShowIcon = false
	cfg.Style.VersionPrefix = "v"
	f := New(cfg)

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
	}

	result := f.Format(versions)

	if !strings.Contains(result, "v18.17.0") {
		t.Errorf("expected version prefix 'v', got '%s'", result)
	}
}

func TestFormat_CustomSeparator(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "plain"
	cfg.Style.ShowIcon = false
	cfg.Style.Separator = " | "
	f := New(cfg)

	versions := []*parser.VersionInfo{
		{Type: "node", Version: "18.17.0", Source: "file", Priority: 10},
		{Type: "go", Version: "1.21", Source: "go.mod", Priority: 20},
	}

	result := f.Format(versions)

	if !strings.Contains(result, " | ") {
		t.Errorf("expected separator ' | ' in output, got '%s'", result)
	}
}

func TestFormatSingle_Tmux(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "tmux"
	f := New(cfg)

	v := &parser.VersionInfo{
		Type:    "node",
		Version: "18.17.0",
		Source:  "file",
	}

	result := f.FormatSingle(v)

	if !strings.Contains(result, "18.17.0") {
		t.Errorf("expected version in output, got '%s'", result)
	}
	if !strings.Contains(result, "#[fg=") {
		t.Errorf("expected tmux color codes, got '%s'", result)
	}
}

func TestFormatSingle_JSON(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "json"
	f := New(cfg)

	v := &parser.VersionInfo{
		Type:    "node",
		Version: "18.17.0",
		Source:  "file",
	}

	result := f.FormatSingle(v)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestTmuxColor(t *testing.T) {
	cfg := config.Default()
	f := New(cfg)

	result := f.tmuxColor("colour39", "test")
	expected := "#[fg=colour39]test#[default]"

	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}

	// Empty color should return text unchanged
	result = f.tmuxColor("", "test")
	if result != "test" {
		t.Errorf("expected 'test' for empty color, got '%s'", result)
	}
}

func TestFormat_AllIcons(t *testing.T) {
	cfg := config.Default()
	cfg.OutputFormat = "plain"
	f := New(cfg)

	types := []struct {
		typ  string
		icon string
	}{
		{"node", "⬢"},
		{"go", "🐹"},
		{"rust", "🦀"},
		{"python", "🐍"},
		{"ruby", "💎"},
		{"php", "🐘"},
		{"java", "☕"},
		{"dotnet", "🔷"},
		{"elixir", "💧"},
		{"deno", "🦕"},
		{"bun", "🥟"},
		{"zig", "⚡"},
		{"swift", "🐦"},
		{"haskell", "λ"},
		{"lua", "🌙"},
		{"perl", "🐪"},
	}

	for _, tc := range types {
		versions := []*parser.VersionInfo{
			{Type: tc.typ, Version: "1.0.0", Source: "file", Priority: 10},
		}

		result := f.Format(versions)

		if !strings.Contains(result, tc.icon) {
			t.Errorf("expected icon '%s' for type '%s', got '%s'", tc.icon, tc.typ, result)
		}
	}
}
