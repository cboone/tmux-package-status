# Plan: tmux-asdf-status TPM Plugin

A TPM plugin that displays asdf tool versions in the tmux status line, similar to p10k's asdf segment.

## File Structure

```
tmux-asdf-status/
├── asdf-status.tmux              # Main TPM entry point (executable)
├── scripts/
│   ├── asdf_status.sh            # Core script generating status output
│   ├── helpers.sh                # Shared helper functions
│   └── variables.sh              # Configuration variable definitions
├── LICENSE                       # (exists)
└── README.md                     # Documentation (update)
```

## How It Works

1. **User adds `#{asdf_status}` to their `status-left` or `status-right`**
2. **TPM loads `asdf-status.tmux`** which replaces the placeholder with:
   ```
   #($PLUGIN_DIR/scripts/asdf_status.sh)
   ```
3. **tmux calls `asdf_status.sh`** which retrieves `pane_current_path` directly via `tmux display-message -p`
4. **Script finds `.tool-versions`** by walking up the directory tree
5. **Script compares local vs global versions** and filters based on config
6. **Script outputs formatted string** with tool names, versions, icons, colors

## Core Algorithm

```
1. Retrieve pane_current_path via `tmux display-message -p "#{pane_current_path}"`
2. Walk up directory tree looking for .tool-versions
3. If not found, exit with empty output
4. Parse local .tool-versions into tool=version pairs
5. Load global ~/.tool-versions for comparison
6. For each local tool:
   - Skip if version matches global (unless @asdf_show_global is on)
   - Skip if version is "system" (unless @asdf_show_system is on)
   - Skip if not in @asdf_tools filter (when set)
   - Format with icon, color, and configured format string
7. Join entries with separator and output
```

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `@asdf_show_global` | `off` | Show tools even when version matches global |
| `@asdf_show_system` | `off` | Show tools set to `system` version |
| `@asdf_tools` | `""` | Comma-separated filter (empty = all tools) |
| `@asdf_separator` | `" "` | Separator between tool entries |
| `@asdf_format` | `"{tool} {version}"` | Format string per tool |
| `@asdf_icon_<tool>` | `""` | Icon for specific tool (e.g., `@asdf_icon_python`) |
| `@asdf_color_<tool>` | `""` | Color for specific tool as `fg[,bg]` |
| `@asdf_default_icon` | `""` | Default icon when tool has no specific icon |
| `@asdf_default_color` | `""` | Default color when tool has no specific color, as `fg[,bg]` |

## Implementation Phases

### Phase 1: Core Functionality
- Create `asdf-status.tmux` with placeholder interpolation
- Create `scripts/helpers.sh` with `get_tmux_option` / `set_tmux_option`
- Create `scripts/variables.sh` with option names and defaults
- Create `scripts/asdf_status.sh` with:
  - Directory walking to find `.tool-versions`
  - File parsing
  - Basic output (tool version pairs)

### Phase 2: Filtering & Comparison
- Add global version loading from `~/.tool-versions`
- Implement `@asdf_show_global` filtering
- Implement `@asdf_show_system` filtering
- Implement `@asdf_tools` whitelist filtering

### Phase 3: Formatting & Styling
- Add `@asdf_format` template support
- Add `@asdf_separator` support
- Add per-tool icon support (`@asdf_icon_<tool>`)
- Add per-tool color support (`@asdf_color_<tool>`)
- Add tmux color syntax (`#[fg=color]...#[default]`)

### Phase 4: Documentation & Polish
- Update README.md with installation and configuration docs
- Add usage examples
- Handle edge cases (permissions, empty files, comments)

## Example User Configuration

```tmux
# Install plugin
set -g @plugin 'your-username/tmux-asdf-status'

# Add to status line
set -g status-right '#{asdf_status} | %H:%M'

# Customize (optional)
set -g @asdf_tools "python,nodejs,ruby"
set -g @asdf_icon_python " "
set -g @asdf_icon_nodejs " "
set -g @asdf_color_python "yellow"
set -g @asdf_format "{version}"
set -g @asdf_separator " | "
```

## Key Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `asdf-status.tmux` | Create | TPM entry point, placeholder replacement |
| `scripts/helpers.sh` | Create | `get_tmux_option`, `set_tmux_option` |
| `scripts/variables.sh` | Create | Option names and default values |
| `scripts/asdf_status.sh` | Create | Core logic for version detection and output |
| `README.md` | Update | Installation and configuration documentation |

## Verification

1. **Install plugin via TPM**: Add to `.tmux.conf` and run `prefix + I`
2. **Test with no `.tool-versions`**: Status should show nothing
3. **Test in directory with `.tool-versions`**: Should display tool versions
4. **Test filtering**: Set `@asdf_show_global off` and verify global-matching tools are hidden
5. **Test formatting**: Configure icons/colors and verify output
6. **Test parent directory detection**: Navigate to subdirectory and verify versions still show
