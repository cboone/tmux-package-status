#!/usr/bin/env bash
# tmux-package-status plugin for TPM (Tmux Plugin Manager)
# This file is the entry point for TPM and TPM-Redux

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Get tmux option or use default
get_tmux_option() {
    local option="$1"
    local default_value="$2"
    local option_value

    option_value="$(tmux show-option -gqv "$option")"
    if [[ -z "$option_value" ]]; then
        echo "$default_value"
    else
        echo "$option_value"
    fi
}

# Set environment variables from tmux options
setup_env_from_tmux_options() {
    # Directory to scan
    local dir
    dir="$(get_tmux_option "@package-status-dir" "")"
    if [[ -n "$dir" ]]; then
        export TMUX_PKG_DIR="$dir"
    fi

    # Output format
    local format
    format="$(get_tmux_option "@package-status-format" "")"
    if [[ -n "$format" ]]; then
        export TMUX_PKG_FORMAT="$format"
    fi

    # Enabled managers
    local managers
    managers="$(get_tmux_option "@package-status-managers" "")"
    if [[ -n "$managers" ]]; then
        export TMUX_PKG_MANAGERS="$managers"
    fi

    # Max items
    local max_items
    max_items="$(get_tmux_option "@package-status-max-items" "")"
    if [[ -n "$max_items" ]]; then
        export TMUX_PKG_MAX_ITEMS="$max_items"
    fi

    # Cache TTL
    local cache_ttl
    cache_ttl="$(get_tmux_option "@package-status-cache-ttl" "")"
    if [[ -n "$cache_ttl" ]]; then
        export TMUX_PKG_CACHE_TTL="$cache_ttl"
    fi

    # Style options
    local icon_color
    icon_color="$(get_tmux_option "@package-status-icon-color" "")"
    if [[ -n "$icon_color" ]]; then
        export TMUX_PKG_ICON_COLOR="$icon_color"
    fi

    local version_color
    version_color="$(get_tmux_option "@package-status-version-color" "")"
    if [[ -n "$version_color" ]]; then
        export TMUX_PKG_VERSION_COLOR="$version_color"
    fi

    local separator_color
    separator_color="$(get_tmux_option "@package-status-separator-color" "")"
    if [[ -n "$separator_color" ]]; then
        export TMUX_PKG_SEPARATOR_COLOR="$separator_color"
    fi

    local show_icons
    show_icons="$(get_tmux_option "@package-status-show-icons" "")"
    if [[ -n "$show_icons" ]]; then
        export TMUX_PKG_SHOW_ICON="$show_icons"
    fi

    local show_brackets
    show_brackets="$(get_tmux_option "@package-status-show-brackets" "")"
    if [[ -n "$show_brackets" ]]; then
        export TMUX_PKG_SHOW_BRACKETS="$show_brackets"
    fi

    local separator
    separator="$(get_tmux_option "@package-status-separator" "")"
    if [[ -n "$separator" ]]; then
        export TMUX_PKG_SEPARATOR="$separator"
    fi

    local prefix
    prefix="$(get_tmux_option "@package-status-prefix" "")"
    if [[ -n "$prefix" ]]; then
        export TMUX_PKG_PREFIX="$prefix"
    fi

    local suffix
    suffix="$(get_tmux_option "@package-status-suffix" "")"
    if [[ -n "$suffix" ]]; then
        export TMUX_PKG_SUFFIX="$suffix"
    fi
}

# Define the format string for tmux
setup_format_string() {
    local script="${CURRENT_DIR}/scripts/tmux-package-status.sh"

    # Make sure the script is executable
    chmod +x "$script" 2>/dev/null || true

    # Set the format string that tmux will use
    # Use #{pane_current_path} to get the directory of the active pane
    tmux set-option -g @package-status "#($script -d '#{pane_current_path}')"
}

# Install binary if not present
install_binary_if_needed() {
    local binary="${CURRENT_DIR}/bin/tmux-package-status"

    # Check if binary exists and is executable
    if [[ -x "$binary" ]]; then
        return 0
    fi

    # Check if install script exists and run it
    local install_script="${CURRENT_DIR}/install.sh"
    if [[ -x "$install_script" ]]; then
        "$install_script" --local 2>/dev/null || true
    fi
}

# Main initialization
main() {
    # Install binary if needed
    install_binary_if_needed

    # Setup environment from tmux options
    setup_env_from_tmux_options

    # Setup the format string
    setup_format_string
}

main
