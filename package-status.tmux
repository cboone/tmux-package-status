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

# Set tmux global environment variables from tmux options
# Using tmux set-environment -g so subprocess commands can access them
setup_env_from_tmux_options() {
    # Helper to set or unset tmux environment variable
    set_tmux_env() {
        local env_name="$1"
        local value="$2"
        if [[ -n "$value" ]]; then
            tmux set-environment -g "$env_name" "$value"
        else
            tmux set-environment -gu "$env_name" 2>/dev/null || true
        fi
    }

    # Directory to scan
    set_tmux_env "TMUX_PKG_DIR" "$(get_tmux_option "@package-status-dir" "")"

    # Output format
    set_tmux_env "TMUX_PKG_FORMAT" "$(get_tmux_option "@package-status-format" "")"

    # Enabled managers
    set_tmux_env "TMUX_PKG_MANAGERS" "$(get_tmux_option "@package-status-managers" "")"

    # Max items
    set_tmux_env "TMUX_PKG_MAX_ITEMS" "$(get_tmux_option "@package-status-max-items" "")"

    # Cache TTL
    set_tmux_env "TMUX_PKG_CACHE_TTL" "$(get_tmux_option "@package-status-cache-ttl" "")"

    # Style options
    set_tmux_env "TMUX_PKG_ICON_COLOR" "$(get_tmux_option "@package-status-icon-color" "")"
    set_tmux_env "TMUX_PKG_VERSION_COLOR" "$(get_tmux_option "@package-status-version-color" "")"
    set_tmux_env "TMUX_PKG_SEPARATOR_COLOR" "$(get_tmux_option "@package-status-separator-color" "")"
    set_tmux_env "TMUX_PKG_SHOW_ICON" "$(get_tmux_option "@package-status-show-icons" "")"
    set_tmux_env "TMUX_PKG_SHOW_BRACKETS" "$(get_tmux_option "@package-status-show-brackets" "")"
    set_tmux_env "TMUX_PKG_SEPARATOR" "$(get_tmux_option "@package-status-separator" "")"
    set_tmux_env "TMUX_PKG_PREFIX" "$(get_tmux_option "@package-status-prefix" "")"
    set_tmux_env "TMUX_PKG_SUFFIX" "$(get_tmux_option "@package-status-suffix" "")"
}

# Define the format string for tmux
setup_format_string() {
    local script="${CURRENT_DIR}/scripts/tmux-package-status.sh"

    # Make sure the script is executable
    chmod +x "$script" 2>/dev/null || true

    # Set the format string that tmux will use
    # The script runs in the pane's current directory via tmux, using $PWD
    # We avoid passing #{pane_current_path} directly to prevent command injection
    # from maliciously crafted directory names
    tmux set-option -g @package-status "#($script)"
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
