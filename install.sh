#!/usr/bin/env bash
set -euo pipefail

# Installs the `mcserver` CLI so it is available from any directory.
# Default: user install (no sudo) into ~/.local/bin via pip.
# Usage:
#   ./install.sh
#   ./install.sh --system   # installs system-wide (requires sudo)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODE="user"

if [[ "${1:-}" == "--system" ]]; then
  MODE="system"
elif [[ "${1:-}" != "" ]]; then
  echo "Usage: $0 [--system]" >&2
  exit 2
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "ERROR: python3 not found" >&2
  exit 1
fi

cd "$ROOT_DIR"

if [[ "$MODE" == "system" ]]; then
  echo "Installing system-wide (sudo)..."
  sudo python3 -m pip install -e .
else
  echo "Installing for current user..."
  python3 -m pip install --user -e .
fi

echo
if command -v mcserver >/dev/null 2>&1; then
  echo "OK: mcserver is installed: $(command -v mcserver)"
else
  echo "Installed, but 'mcserver' is not on PATH yet." >&2
  echo "On Linux, user installs typically go to: ~/.local/bin" >&2
  echo "Add it to PATH, e.g.:" >&2
  echo "  echo 'export PATH=\"$HOME/.local/bin:$PATH\"' >> ~/.bashrc" >&2
  echo "  source ~/.bashrc" >&2
  exit 1
fi

echo
echo "Next: save your CurseForge API key:" 
echo "  mcserver config set-api-key"
echo "Try:" 
echo "  mcserver --help"
