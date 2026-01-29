#!/usr/bin/env bash
# tmux-package-status installation script
# Usage: curl -fsSL https://raw.githubusercontent.com/cboone/tmux-package-status/main/install.sh | bash

set -e

# Configuration
GITHUB_REPO="cboone/tmux-package-status"
BINARY_NAME="tmux-package-status"
DEFAULT_INSTALL_DIR="${HOME}/.local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored message
print_color() {
    local color="$1"
    local message="$2"
    printf "${color}%s${NC}\n" "$message"
}

# Print step
print_step() {
    printf "${BLUE}==>${NC} %s\n" "$1"
}

# Print success
print_success() {
    printf "${GREEN}✓${NC} %s\n" "$1"
}

# Print warning
print_warning() {
    printf "${YELLOW}!${NC} %s\n" "$1"
}

# Print error and exit
print_error() {
    printf "${RED}✗${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*)
            echo "darwin"
            ;;
        Linux*)
            echo "linux"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            echo "windows"
            ;;
        FreeBSD*)
            echo "freebsd"
            ;;
        *)
            print_error "Unsupported operating system: $(uname -s)"
            ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)
            echo "amd64"
            ;;
        arm64|aarch64)
            echo "arm64"
            ;;
        armv7l|armv6l)
            echo "arm"
            ;;
        i386|i686)
            echo "386"
            ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            ;;
    esac
}

# Get the latest release version from GitHub
get_latest_version() {
    local url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    local version

    if command -v curl &>/dev/null; then
        version=$(curl -fsSL "$url" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget &>/dev/null; then
        version=$(wget -qO- "$url" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget found. Please install one of them."
    fi

    if [[ -z "$version" ]]; then
        print_error "Failed to get latest version from GitHub"
    fi

    echo "$version"
}

# Download and install the binary
download_and_install() {
    local version="$1"
    local os="$2"
    local arch="$3"
    local install_dir="$4"

    # Construct download URL
    # Windows uses .zip archives, others use .tar.gz
    local archive_ext="tar.gz"
    if [[ "$os" == "windows" ]]; then
        archive_ext="zip"
    fi
    local archive_name="${BINARY_NAME}_${version#v}_${os}_${arch}.${archive_ext}"
    local url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${archive_name}"

    # Create temp directory
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf '$tmp_dir'" EXIT

    print_step "Downloading ${BINARY_NAME} ${version} for ${os}/${arch}..."

    # Download archive
    local archive_path="${tmp_dir}/${archive_name}"
    if command -v curl &>/dev/null; then
        if ! curl -fsSL -o "$archive_path" "$url" 2>/dev/null; then
            print_error "Failed to download from $url"
        fi
    elif command -v wget &>/dev/null; then
        if ! wget -q -O "$archive_path" "$url" 2>/dev/null; then
            print_error "Failed to download from $url"
        fi
    fi

    print_step "Extracting archive..."

    # Extract archive based on type
    if [[ "$os" == "windows" ]]; then
        if command -v unzip &>/dev/null; then
            unzip -q "$archive_path" -d "$tmp_dir"
        else
            print_error "unzip is required to extract Windows archives"
        fi
    else
        tar -xzf "$archive_path" -C "$tmp_dir"
    fi

    # Create install directory if needed
    mkdir -p "$install_dir"

    # Install binary (Windows uses .exe extension)
    local binary_name="${BINARY_NAME}"
    local install_name="${BINARY_NAME}"
    if [[ "$os" == "windows" ]]; then
        binary_name="${BINARY_NAME}.exe"
        install_name="${BINARY_NAME}.exe"
    fi

    local binary_path="${tmp_dir}/${binary_name}"

    if [[ ! -f "$binary_path" ]]; then
        print_error "Binary not found in archive"
    fi

    chmod +x "$binary_path"
    mv "$binary_path" "${install_dir}/${install_name}"

    print_success "Installed to ${install_dir}/${install_name}"
}

# Add to PATH if needed
setup_path() {
    local install_dir="$1"

    # Check if already in PATH
    if echo "$PATH" | tr ':' '\n' | grep -qx "$install_dir"; then
        return 0
    fi

    print_warning "$install_dir is not in your PATH"
    echo ""
    echo "Add the following to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo ""
    echo "  export PATH=\"\$PATH:$install_dir\""
    echo ""
}

# Setup tmux configuration
print_tmux_config() {
    echo ""
    print_step "Add the following to your tmux.conf:"
    echo ""
    echo "  # Using TPM-Redux (recommended):"
    echo "  set -g @plugin 'cboone/tmux-package-status'"
    echo ""
    echo "  # Or manually:"
    echo "  set -g status-right '#{@package-status}'"
    echo ""
    echo "  # Optional configuration:"
    echo "  set -g @package-status-max-items 3"
    echo "  set -g @package-status-icon-color 'colour39'"
    echo "  set -g @package-status-show-brackets true"
    echo ""
}

# Show usage
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Options:
    -d, --dir DIR      Install to DIR (default: ${DEFAULT_INSTALL_DIR})
    -v, --version VER  Install specific version (default: latest)
    -l, --local        Install to ./bin (for TPM plugin use)
    -h, --help         Show this help message

Examples:
    # Install latest version to ~/.local/bin
    $0

    # Install specific version
    $0 -v v1.0.0

    # Install to custom directory
    $0 -d /usr/local/bin

    # Install locally (for TPM plugin)
    $0 --local
EOF
}

# Parse command line arguments
parse_args() {
    INSTALL_DIR="$DEFAULT_INSTALL_DIR"
    VERSION=""
    LOCAL_INSTALL=false

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -l|--local)
                LOCAL_INSTALL=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                ;;
        esac
    done

    if $LOCAL_INSTALL; then
        SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        INSTALL_DIR="${SCRIPT_DIR}/bin"
    fi
}

# Main installation
main() {
    parse_args "$@"

    echo ""
    print_color "$BLUE" "tmux-package-status installer"
    echo ""

    # Detect platform
    local os arch
    os=$(detect_os)
    arch=$(detect_arch)

    print_step "Detected platform: ${os}/${arch}"

    # Get version
    if [[ -z "$VERSION" ]]; then
        print_step "Fetching latest version..."
        VERSION=$(get_latest_version)
    fi

    print_step "Version: ${VERSION}"

    # Download and install
    download_and_install "$VERSION" "$os" "$arch" "$INSTALL_DIR"

    # Setup PATH if not local install
    if ! $LOCAL_INSTALL; then
        setup_path "$INSTALL_DIR"
    fi

    # Print tmux configuration
    print_tmux_config

    print_success "Installation complete!"
}

main "$@"
