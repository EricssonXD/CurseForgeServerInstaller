#!/bin/sh
set -e

# CurseForge Server Installer (cfs) - one-line installer
# Usage: curl -fsSL https://raw.githubusercontent.com/EricssonXD/CurseForgeServerInstaller/master/install-cfs.sh | sh

REPO="EricssonXD/CurseForgeServerInstaller"
INSTALL_DIR="${CFS_INSTALL_DIR:-/usr/local/bin}"

# Detect OS and arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
    linux|darwin) ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Get latest release tag
TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
if [ -z "$TAG" ]; then
    echo "Error: could not determine latest release" >&2
    exit 1
fi

VERSION="${TAG#v}"
EXT="tar.gz"
[ "$OS" = "windows" ] && EXT="zip"

FILENAME="cfs_${VERSION}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${TAG}/${FILENAME}"

echo "Downloading cfs ${TAG} for ${OS}/${ARCH}..."
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "${TMP}/${FILENAME}"

if [ "$EXT" = "tar.gz" ]; then
    tar -xzf "${TMP}/${FILENAME}" -C "$TMP"
else
    unzip -q "${TMP}/${FILENAME}" -d "$TMP"
fi

# Install binary
if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP}/cfs" "${INSTALL_DIR}/cfs"
else
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "${TMP}/cfs" "${INSTALL_DIR}/cfs"
fi
chmod +x "${INSTALL_DIR}/cfs"

echo "cfs ${TAG} installed to ${INSTALL_DIR}/cfs"

# Set up shell completion
SHELL_NAME="$(basename "$SHELL" 2>/dev/null || echo "")"
case "$SHELL_NAME" in
    bash)
        COMP_LINE='eval "$(cfs completion bash)"'
        RC="$HOME/.bashrc"
        if [ -f "$RC" ] && ! grep -qF 'cfs completion bash' "$RC"; then
            echo "$COMP_LINE" >> "$RC"
            echo "Added bash completion to $RC"
        fi
        ;;
    zsh)
        COMP_LINE='eval "$(cfs completion zsh)"'
        RC="$HOME/.zshrc"
        if [ -f "$RC" ] && ! grep -qF 'cfs completion zsh' "$RC"; then
            echo "$COMP_LINE" >> "$RC"
            echo "Added zsh completion to $RC"
        fi
        ;;
    fish)
        COMP_DIR="$HOME/.config/fish/completions"
        mkdir -p "$COMP_DIR"
        cfs completion fish > "$COMP_DIR/cfs.fish"
        echo "Added fish completion to $COMP_DIR/cfs.fish"
        ;;
esac
echo "Run 'cfs --help' to get started"
