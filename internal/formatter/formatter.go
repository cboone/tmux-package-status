// Package formatter formats version information for tmux status bar
package formatter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cboone/tmux-package-status/internal/config"
	"github.com/cboone/tmux-package-status/internal/parser"
)

// Formatter formats version info for output
type Formatter struct {
	cfg *config.Config
}

// New creates a new Formatter with the given configuration
func New(cfg *config.Config) *Formatter {
	return &Formatter{cfg: cfg}
}

// Format formats version information according to the output format
func (f *Formatter) Format(versions []*parser.VersionInfo) string {
	switch f.cfg.OutputFormat {
	case "json":
		return f.formatJSON(versions)
	case "plain":
		return f.formatPlain(versions)
	default:
		return f.formatTmux(versions)
	}
}

// formatTmux formats output with tmux styling
func (f *Formatter) formatTmux(versions []*parser.VersionInfo) string {
	if len(versions) == 0 {
		return ""
	}

	// Apply max items limit
	if f.cfg.MaxItems > 0 && len(versions) > f.cfg.MaxItems {
		versions = versions[:f.cfg.MaxItems]
	}

	var parts []string
	for _, v := range versions {
		part := f.formatSingleTmux(v)
		if part != "" {
			parts = append(parts, part)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	// Join with separator
	sep := f.tmuxColor(f.cfg.Style.SeparatorColor, f.cfg.Style.Separator)
	result := strings.Join(parts, sep)

	// Add prefix/suffix
	if f.cfg.Style.Prefix != "" {
		result = f.cfg.Style.Prefix + result
	}
	if f.cfg.Style.Suffix != "" {
		result = result + f.cfg.Style.Suffix
	}

	return result
}

// formatSingleTmux formats a single version entry for tmux
func (f *Formatter) formatSingleTmux(v *parser.VersionInfo) string {
	var parts []string

	// Icon
	if f.cfg.Style.ShowIcon {
		icon := f.cfg.Style.Icons[v.Type]
		if icon != "" {
			iconStr := f.tmuxColor(f.cfg.Style.IconColor, icon)
			parts = append(parts, iconStr)
		}
	}

	// Version with optional prefix
	version := f.cfg.Style.VersionPrefix + v.Version
	versionStr := f.tmuxColor(f.cfg.Style.VersionColor, version)

	// Brackets
	if f.cfg.Style.ShowBrackets {
		versionStr = f.tmuxColor(f.cfg.Style.BracketColor, "[") +
			versionStr +
			f.tmuxColor(f.cfg.Style.BracketColor, "]")
	}

	parts = append(parts, versionStr)

	return strings.Join(parts, " ")
}

// tmuxColor wraps text in tmux color formatting
func (f *Formatter) tmuxColor(color, text string) string {
	if color == "" {
		return text
	}
	return fmt.Sprintf("#[fg=%s]%s#[default]", color, text)
}

// formatPlain formats output without any styling
func (f *Formatter) formatPlain(versions []*parser.VersionInfo) string {
	if len(versions) == 0 {
		return ""
	}

	// Apply max items limit
	if f.cfg.MaxItems > 0 && len(versions) > f.cfg.MaxItems {
		versions = versions[:f.cfg.MaxItems]
	}

	var parts []string
	for _, v := range versions {
		part := f.formatSinglePlain(v)
		if part != "" {
			parts = append(parts, part)
		}
	}

	result := strings.Join(parts, f.cfg.Style.Separator)

	if f.cfg.Style.Prefix != "" {
		result = f.cfg.Style.Prefix + result
	}
	if f.cfg.Style.Suffix != "" {
		result = result + f.cfg.Style.Suffix
	}

	return result
}

// formatSinglePlain formats a single version entry without styling
func (f *Formatter) formatSinglePlain(v *parser.VersionInfo) string {
	var parts []string

	if f.cfg.Style.ShowIcon {
		icon := f.cfg.Style.Icons[v.Type]
		if icon != "" {
			parts = append(parts, icon)
		}
	}

	version := f.cfg.Style.VersionPrefix + v.Version

	if f.cfg.Style.ShowBrackets {
		version = "[" + version + "]"
	}

	parts = append(parts, version)

	return strings.Join(parts, " ")
}

// formatJSON formats output as JSON
func (f *Formatter) formatJSON(versions []*parser.VersionInfo) string {
	type jsonVersion struct {
		Type    string `json:"type"`
		Version string `json:"version"`
		Source  string `json:"source"`
	}

	// Apply max items limit
	if f.cfg.MaxItems > 0 && len(versions) > f.cfg.MaxItems {
		versions = versions[:f.cfg.MaxItems]
	}

	jv := make([]jsonVersion, len(versions))
	for i, v := range versions {
		jv[i] = jsonVersion{
			Type:    v.Type,
			Version: v.Version,
			Source:  v.Source,
		}
	}

	data, err := json.Marshal(jv)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// FormatSingle formats a single version for display (useful for testing)
func (f *Formatter) FormatSingle(v *parser.VersionInfo) string {
	switch f.cfg.OutputFormat {
	case "json":
		data, _ := json.Marshal(v)
		return string(data)
	case "plain":
		return f.formatSinglePlain(v)
	default:
		return f.formatSingleTmux(v)
	}
}
