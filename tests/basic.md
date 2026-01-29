# Basic Tests

These tests verify the basic functionality of tmux-package-status.

## Help

The help flag should display usage information.

```console
$ tmux-package-status --help
tmux-package-status - Display package version information for tmux status bar

USAGE:
  tmux-package-status [OPTIONS]

OPTIONS:
  -d, --dir <path>         Directory to scan (default: current directory)
  -f, --format <format>    Output format: tmux, plain, json (default: tmux)
  -m, --managers <list>    Comma-separated list of package managers to show
      --max <n>            Maximum items to show (default: 5)
      --cache-ttl <sec>    Cache TTL in seconds (default: 30)
      --no-cache           Disable caching
      --clear-cache        Clear cache and exit

  Style Options:
      --icon-color <color>      Icon color (tmux format, e.g., colour39)
      --version-color <color>   Version color
      --separator-color <color> Separator color
      --bracket-color <color>   Bracket color
      --icons                   Show icons (default: true)
      --brackets                Show brackets around versions
      --separator <sep>         Separator between items (default: space)
      --prefix <prefix>         Prefix for output
      --suffix <suffix>         Suffix for output
      --version-prefix <pre>    Prefix for version numbers (e.g., 'v')

  Other:
  -v, --version            Show version information
  -h, --help               Show this help message
      --list-managers      List supported package managers

ENVIRONMENT VARIABLES:
  TMUX_PKG_DIR             Directory to scan
  TMUX_PKG_FORMAT          Output format
  TMUX_PKG_MANAGERS        Comma-separated list of managers
  TMUX_PKG_MAX_ITEMS       Maximum items to show
  TMUX_PKG_CACHE_TTL       Cache TTL in seconds
  TMUX_PKG_ICON_COLOR      Icon color
  TMUX_PKG_VERSION_COLOR   Version color
  TMUX_PKG_SHOW_ICON       Show icons (true/false)
  TMUX_PKG_SHOW_BRACKETS   Show brackets (true/false)
  TMUX_PKG_SEPARATOR       Separator between items
  TMUX_PKG_PREFIX          Output prefix
  TMUX_PKG_SUFFIX          Output suffix

EXAMPLES:
  # Basic usage in tmux.conf:
  set -g status-right '#{E:tmux_package_status}'

  # Show only Node.js and Go versions:
  tmux-package-status -m node,go

  # Plain text output:
  tmux-package-status -f plain

  # JSON output:
  tmux-package-status -f json

```

## Version

The version flag should display version information.

```console
$ tmux-package-status --version
tmux-package-status dev
  commit: none
  built:  unknown

```

## List Managers

The list-managers flag should display all supported package managers.

```console
$ tmux-package-status --list-managers
Supported package managers:

  ⬢ node
    Files: package.json, .nvmrc, .node-version
  🐹 go
    Files: go.mod
  🦀 rust
    Files: Cargo.toml, rust-toolchain.toml
  🐍 python
    Files: pyproject.toml, requirements.txt, .python-version, Pipfile
  💎 ruby
    Files: Gemfile, .ruby-version
  🐘 php
    Files: composer.json
  ☕ java
    Files: pom.xml, build.gradle
  🔷 dotnet
    Files: *.csproj, *.fsproj, global.json
  💧 elixir
    Files: mix.exs
  🦕 deno
    Files: deno.json, deno.jsonc
  🥟 bun
    Files: bun.lockb, bunfig.toml
  ⚡ zig
    Files: build.zig, build.zig.zon
  🐦 swift
    Files: Package.swift
  K kotlin
    Files: build.gradle.kts
  S scala
    Files: build.sbt
  λ haskell
    Files: stack.yaml, *.cabal
  λ clojure
    Files: project.clj, deps.edn
  🌙 lua
    Files: *.rockspec
  🐪 perl
    Files: cpanfile, Makefile.PL

```

## Empty Directory

Scanning an empty directory should produce no output.

```console
$ mkdir -p /tmp/empty-test-dir && tmux-package-status -d /tmp/empty-test-dir --no-cache

```

## Clear Cache

The clear-cache flag should clear the cache successfully.

```console
$ tmux-package-status --clear-cache
Cache cleared

```
