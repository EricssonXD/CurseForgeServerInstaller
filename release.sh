#!/bin/sh
set -e

# Release helper for cfs
# Usage:
#   ./release.sh patch    # 0.1.0 -> 0.1.1
#   ./release.sh minor    # 0.1.0 -> 0.2.0
#   ./release.sh major    # 0.1.0 -> 1.0.0
#   ./release.sh          # shows current version

VERSION_FILE="VERSION"
CURRENT="$(cat "$VERSION_FILE" | tr -d '[:space:]')"

if [ -z "$1" ]; then
    echo "Current version: v${CURRENT}"
    exit 0
fi

BUMP="$1"
IFS='.' read -r MAJOR MINOR PATCH <<EOF
$CURRENT
EOF

case "$BUMP" in
    patch) PATCH=$((PATCH + 1)) ;;
    minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
    major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
    *)
        echo "Usage: $0 [patch|minor|major]" >&2
        exit 1
        ;;
esac

NEW="${MAJOR}.${MINOR}.${PATCH}"
TAG="v${NEW}"

echo "Bumping: v${CURRENT} -> ${TAG}"

# Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: uncommitted changes. Commit or stash first." >&2
    exit 1
fi

# Update VERSION file
printf '%s\n' "$NEW" > "$VERSION_FILE"

# Commit and tag
git add "$VERSION_FILE"
git commit -m "release: ${TAG}"
git tag "$TAG"

echo ""
echo "Created tag ${TAG}. Push with:"
echo "  git push && git push origin ${TAG}"
echo ""
echo "This will trigger the GitHub Actions release workflow."
