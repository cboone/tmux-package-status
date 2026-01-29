#!/usr/bin/env bash
# tmux-package-status wrapper script
# This script detects the platform and runs the appropriate binary

set -e

# Determine the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_DIR="$(dirname "$SCRIPT_DIR")"

# Detect OS and architecture
detect_platform() {
    local os arch

    case "$(uname -s)" in
        Darwin*)
            os="darwin"
            ;;
        Linux*)
            os="linux"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            os="windows"
            ;;
        FreeBSD*)
            os="freebsd"
            ;;
        *)
            echo "Unsupported operating system: $(uname -s)" >&2
            exit 1
            ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        arm64|aarch64)
            arch="arm64"
            ;;
        armv7l|armv6l)
            arch="arm"
            ;;
        i386|i686)
            arch="386"
            ;;
        *)
            echo "Unsupported architecture: $(uname -m)" >&2
            exit 1
            ;;
    esac

    echo "${os}_${arch}"
}

# Find the binary
find_binary() {
    local platform="$1"
    local binary_name="tmux-package-status"

    # Add .exe extension on Windows
    if [[ "$platform" == windows_* ]]; then
        binary_name="${binary_name}.exe"
    fi

    # Check multiple possible locations
    local locations=(
        "${PLUGIN_DIR}/bin/${binary_name}"
        "${PLUGIN_DIR}/bin/${platform}/${binary_name}"
        "${HOME}/.local/bin/${binary_name}"
        "/usr/local/bin/${binary_name}"
        "/usr/bin/${binary_name}"
    )

    for loc in "${locations[@]}"; do
        if [[ -x "$loc" ]]; then
            echo "$loc"
            return 0
        fi
    done

    # Try to find it in PATH
    if command -v "$binary_name" &>/dev/null; then
        command -v "$binary_name"
        return 0
    fi

    return 1
}

# Main
main() {
    local platform
    platform="$(detect_platform)"

    local binary
    if ! binary="$(find_binary "$platform")"; then
        echo "Error: tmux-package-status binary not found" >&2
        echo "Please install using one of the following methods:" >&2
        echo "  1. TPM: Add 'set -g @plugin \"cboone/tmux-package-status\"' to tmux.conf" >&2
        echo "  2. Manual: curl -fsSL https://raw.githubusercontent.com/cboone/tmux-package-status/main/install.sh | sh" >&2
        exit 1
    fi

    # Execute the binary with all passed arguments
    exec "$binary" "$@"
}

main "$@"
