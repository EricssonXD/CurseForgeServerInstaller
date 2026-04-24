#!/bin/sh
# setup-dev.sh — sets up the dev environment for CurseForge Server Installer
set -e

REPO_ROOT="$(cd "$(dirname "$0")" && pwd)"
GO_MIN="1.22"

echo "=== CurseForge Server Installer — Dev Setup ==="
echo

# ── Check Go ────────────────────────────────────────────────────────────────
if ! command -v go >/dev/null 2>&1; then
    # Try the common non-PATH location
    if [ -x /usr/local/go/bin/go ]; then
        export PATH="$PATH:/usr/local/go/bin"
    else
        echo "ERROR: Go not found. Install Go $GO_MIN+ from https://go.dev/dl/" >&2
        exit 1
    fi
fi

GO_VER="$(go version | awk '{print $3}' | sed 's/go//')"
echo "✓ Go $GO_VER found at $(command -v go)"

# ── Download dependencies ────────────────────────────────────────────────────
echo
echo "→ Downloading dependencies..."
cd "$REPO_ROOT"
go mod download
go mod tidy
echo "✓ Dependencies ready"

# ── Build ────────────────────────────────────────────────────────────────────
echo
echo "→ Building cfs..."
go build -o cfs ./cmd/cfs
echo "✓ Built ./cfs"

# ── Run tests ────────────────────────────────────────────────────────────────
echo
echo "→ Running tests..."
go test ./...
echo "✓ All tests pass"

# ── Shell completion ─────────────────────────────────────────────────────────
echo
echo "→ Setting up shell completion (optional)..."
SHELL_NAME="$(basename "$SHELL" 2>/dev/null || echo "")"
case "$SHELL_NAME" in
    bash)
        RC="$HOME/.bashrc"
        if [ -f "$RC" ] && ! grep -qF 'cfs completion bash' "$RC"; then
            echo 'eval "$('"$REPO_ROOT"'/cfs completion bash)"' >> "$RC"
            echo "  Added bash completion to $RC"
        else
            echo "  Bash completion already configured (or $RC not found)"
        fi
        ;;
    zsh)
        RC="$HOME/.zshrc"
        if [ -f "$RC" ] && ! grep -qF 'cfs completion zsh' "$RC"; then
            echo 'eval "$('"$REPO_ROOT"'/cfs completion zsh)"' >> "$RC"
            echo "  Added zsh completion to $RC"
        else
            echo "  Zsh completion already configured (or $RC not found)"
        fi
        ;;
    fish)
        COMP_DIR="$HOME/.config/fish/completions"
        mkdir -p "$COMP_DIR"
        "$REPO_ROOT/cfs" completion fish > "$COMP_DIR/cfs.fish"
        echo "  Fish completion written to $COMP_DIR/cfs.fish"
        ;;
    *)
        echo "  Shell '$SHELL_NAME' not recognised — run 'cfs completion --help' to set up manually"
        ;;
esac

# ── git hooks (optional) ─────────────────────────────────────────────────────
if [ -d "$REPO_ROOT/.git" ]; then
    HOOK="$REPO_ROOT/.git/hooks/pre-commit"
    if [ ! -f "$HOOK" ]; then
        cat > "$HOOK" <<'HOOK_SCRIPT'
#!/bin/sh
export PATH="$PATH:/usr/local/go/bin"
go build ./... || exit 1
go test ./... || exit 1
HOOK_SCRIPT
        chmod +x "$HOOK"
        echo
        echo "✓ Installed pre-commit hook (build + test)"
    else
        echo
        echo "  pre-commit hook already exists, skipping"
    fi
fi

# ── Done ─────────────────────────────────────────────────────────────────────
echo
echo "=== Dev setup complete ==="
echo "  Binary:    $REPO_ROOT/cfs"
echo "  Run tests: go test ./..."
echo "  Build:     go build -o cfs ./cmd/cfs"
echo "  Release:   ./release.sh [patch|minor|major]"
