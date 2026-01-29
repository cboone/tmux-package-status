# tmux-package-status

Display package version information in your tmux status bar. Automatically detects package managers in your current project directory and shows the relevant language/tool versions.

![tmux-package-status screenshot](https://via.placeholder.com/800x100.png?text=⬢+18.17.0+🐹+1.21+🦀+1.75)

## Features

- **Auto-detection**: Automatically detects package files in your project
- **20+ Package Managers**: Node.js, Go, Rust, Python, Ruby, PHP, Java, and more
- **Customizable**: Colors, icons, separators, and more
- **Fast**: Written in Go with caching support
- **Multiple Output Formats**: tmux, plain text, or JSON
- **Easy Installation**: TPM-Redux, manual, or curl installer

## Supported Package Managers

| Language | Files Detected | Icon |
|----------|---------------|------|
| Node.js | `package.json`, `.nvmrc`, `.node-version` | ⬢ |
| Go | `go.mod` | 🐹 |
| Rust | `Cargo.toml`, `rust-toolchain.toml` | 🦀 |
| Python | `pyproject.toml`, `.python-version`, `Pipfile` | 🐍 |
| Ruby | `Gemfile`, `.ruby-version` | 💎 |
| PHP | `composer.json` | 🐘 |
| Java | `pom.xml`, `build.gradle` | ☕ |
| .NET | `*.csproj`, `global.json` | 🔷 |
| Elixir | `mix.exs` | 💧 |
| Deno | `deno.json` | 🦕 |
| Bun | `bun.lockb`, `bunfig.toml` | 🥟 |
| Zig | `build.zig`, `build.zig.zon` | ⚡ |
| Swift | `Package.swift` | 🐦 |
| Kotlin | `build.gradle.kts` | K |
| Scala | `build.sbt` | S |
| Haskell | `stack.yaml`, `*.cabal` | λ |
| Clojure | `project.clj`, `deps.edn` | λ |
| Lua | `*.rockspec` | 🌙 |
| Perl | `cpanfile`, `Makefile.PL` | 🐪 |

## Installation

### Using TPM-Redux (Recommended)

Add this to your `~/.tmux.conf`:

```bash
# Install TPM-Redux first if you haven't:
# git clone https://github.com/tmux-plugins/tpm-redux ~/.tmux/plugins/tpm-redux

set -g @plugin 'cboone/tmux-package-status'

# Initialize TPM-Redux (keep at bottom of tmux.conf)
run '~/.tmux/plugins/tpm-redux/tpm'
```

Then press `prefix + I` to install.

### Using the Installer Script

```bash
curl -fsSL https://raw.githubusercontent.com/cboone/tmux-package-status/main/install.sh | bash
```

### Manual Installation

1. Download the latest release for your platform:

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/cboone/tmux-package-status/releases/latest/download/tmux-package-status_darwin_arm64.tar.gz
tar -xzf tmux-package-status_darwin_arm64.tar.gz

# macOS (Intel)
curl -LO https://github.com/cboone/tmux-package-status/releases/latest/download/tmux-package-status_darwin_amd64.tar.gz
tar -xzf tmux-package-status_darwin_amd64.tar.gz

# Linux (x86_64)
curl -LO https://github.com/cboone/tmux-package-status/releases/latest/download/tmux-package-status_linux_amd64.tar.gz
tar -xzf tmux-package-status_linux_amd64.tar.gz

# Linux (ARM64)
curl -LO https://github.com/cboone/tmux-package-status/releases/latest/download/tmux-package-status_linux_arm64.tar.gz
tar -xzf tmux-package-status_linux_arm64.tar.gz
```

2. Move to a directory in your PATH:

```bash
mv tmux-package-status ~/.local/bin/
chmod +x ~/.local/bin/tmux-package-status
```

3. Add to your `~/.tmux.conf`:

```bash
set -g status-right '#(tmux-package-status -d "#{pane_current_path}")'
```

4. Reload tmux configuration:

```bash
tmux source-file ~/.tmux.conf
```

## Usage

### Basic Usage

The plugin automatically displays version information in your status bar based on the current pane's directory:

```bash
# Default usage (in tmux.conf)
set -g status-right '#(tmux-package-status -d "#{pane_current_path}")'
```

### Command Line

```bash
# Scan current directory
tmux-package-status

# Scan specific directory
tmux-package-status -d /path/to/project

# Output formats
tmux-package-status -f plain    # No tmux formatting
tmux-package-status -f json     # JSON output

# Show only specific managers
tmux-package-status -m node,go,rust

# Limit items shown
tmux-package-status --max 3
```

## Configuration

### Tmux Options

When using TPM-Redux, you can configure the plugin with tmux options:

```bash
# Directory to scan (defaults to current pane path)
set -g @package-status-dir ""

# Maximum items to show
set -g @package-status-max-items 5

# Cache TTL in seconds (0 to disable)
set -g @package-status-cache-ttl 30

# Colors (tmux format)
set -g @package-status-icon-color "colour39"
set -g @package-status-version-color "colour250"
set -g @package-status-separator-color "colour240"

# Show/hide options
set -g @package-status-show-icons "true"
set -g @package-status-show-brackets "false"

# Custom separator between items
set -g @package-status-separator " "

# Prefix/suffix for entire output
set -g @package-status-prefix ""
set -g @package-status-suffix ""

# Filter to specific managers
set -g @package-status-managers ""  # comma-separated, empty = all
```

### Environment Variables

You can also configure via environment variables:

```bash
export TMUX_PKG_DIR="/path/to/project"
export TMUX_PKG_FORMAT="tmux"          # tmux, plain, json
export TMUX_PKG_MANAGERS="node,go"     # comma-separated
export TMUX_PKG_MAX_ITEMS=5
export TMUX_PKG_CACHE_TTL=30

# Styling
export TMUX_PKG_ICON_COLOR="colour39"
export TMUX_PKG_VERSION_COLOR="colour250"
export TMUX_PKG_SEPARATOR_COLOR="colour240"
export TMUX_PKG_SHOW_ICON="true"
export TMUX_PKG_SHOW_BRACKETS="false"
export TMUX_PKG_SEPARATOR=" "
export TMUX_PKG_PREFIX=""
export TMUX_PKG_SUFFIX=""
export TMUX_PKG_VERSION_PREFIX=""      # e.g., "v" for v18.17.0
```

### Command Line Flags

```
Usage: tmux-package-status [OPTIONS]

Options:
  -d, --dir <path>         Directory to scan (default: current directory)
  -f, --format <format>    Output format: tmux, plain, json (default: tmux)
  -m, --managers <list>    Comma-separated list of package managers to show
      --max <n>            Maximum items to show (default: 5)
      --cache-ttl <sec>    Cache TTL in seconds (default: 30)
      --no-cache           Disable caching
      --clear-cache        Clear cache and exit

Style Options:
      --icon-color <color>      Icon color (tmux format)
      --version-color <color>   Version color
      --separator-color <color> Separator color
      --bracket-color <color>   Bracket color
      --icons                   Show icons (default: true)
      --brackets                Show brackets around versions
      --separator <sep>         Separator between items
      --prefix <prefix>         Prefix for output
      --suffix <suffix>         Suffix for output
      --version-prefix <pre>    Prefix for version numbers

Other:
  -v, --version            Show version information
  -h, --help               Show help message
      --list-managers      List supported package managers
```

## Examples

### Basic Status Bar

```bash
set -g status-right '#{@package-status}'
```

Output: `⬢ 18.17.0 🐹 1.21`

### With Brackets

```bash
set -g @package-status-show-brackets "true"
```

Output: `⬢ [18.17.0] 🐹 [1.21]`

### Custom Colors

```bash
set -g @package-status-icon-color "colour208"    # Orange icons
set -g @package-status-version-color "colour46"  # Green versions
```

### Pipe Separator

```bash
set -g @package-status-separator " | "
```

Output: `⬢ 18.17.0 | 🐹 1.21`

### Only Node and Python

```bash
set -g @package-status-managers "node,python"
```

### JSON for Scripts

```bash
$ tmux-package-status -f json
[{"type":"node","version":"18.17.0","source":"file:.nvmrc"},{"type":"go","version":"1.21","source":"go.mod"}]
```

### Integration with Powerline/Airline

```bash
# Add to your tmux.conf
set -g status-right '#[fg=colour39]#(tmux-package-status -d "#{pane_current_path}" -f plain)'
```

## Inspiration

This plugin was inspired by:

- [Oh My Zsh](https://ohmyz.sh/) - Prompt segments showing tool versions
- [asdf](https://asdf-vm.com/) - Universal version manager
- [Starship](https://starship.rs/) - Cross-shell prompt with version info
- [Powerline](https://github.com/powerline/powerline) - Status line plugin
- [tmux-battery](https://github.com/tmux-plugins/tmux-battery) - TPM plugin pattern

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/cboone/tmux-package-status.git
cd tmux-package-status

# Build
go build -o tmux-package-status ./cmd/tmux-package-status/

# Run tests
go test ./...

# Run integration tests (requires Scrut)
pip install scrut
scrut test tests/*.md
```

### Project Structure

```
tmux-package-status/
├── cmd/tmux-package-status/    # Main CLI application
├── internal/
│   ├── cache/                  # Caching functionality
│   ├── config/                 # Configuration handling
│   ├── detector/               # Package file detection
│   ├── formatter/              # Output formatting
│   └── parser/                 # Version parsing
├── scripts/                    # Shell scripts
├── tests/                      # Integration tests (Scrut)
├── .github/workflows/          # CI/CD
├── install.sh                  # Installation script
├── package-status.tmux         # TPM plugin entry point
└── README.md
```

### Running Tests

```bash
# Unit tests
go test -v ./...

# With coverage
go test -v -cover ./...

# Integration tests
scrut test tests/*.md
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- The tmux community for the plugin ecosystem
- All the language version managers this tool integrates with
