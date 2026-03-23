#!/usr/bin/env bash
set -euo pipefail

# Release helper for toks-go SDK
# Usage: ./scripts/release.sh v0.2.0

VERSION="${1:?Usage: ./scripts/release.sh <version> (e.g. v0.2.0)}"

# Validate semver format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: Invalid semver format: $VERSION (expected vX.Y.Z)"
  exit 1
fi

# Must be on main
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$BRANCH" != "main" ]]; then
  echo "Error: Must be on main branch (currently: $BRANCH)"
  exit 1
fi

# Must be up to date with remote
git fetch origin main
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/main)
if [[ "$LOCAL" != "$REMOTE" ]]; then
  echo "Error: Local main is not up to date with origin/main"
  echo "  Local:  $LOCAL"
  echo "  Remote: $REMOTE"
  echo "Run: git pull origin main"
  exit 1
fi

# Tag must not already exist
if git rev-parse "$VERSION" >/dev/null 2>&1; then
  echo "Error: Tag $VERSION already exists"
  exit 1
fi

# Quality gates
echo "Running quality gates..."
echo "  go vet..."
go vet ./...
echo "  go test..."
go test -race -count=1 ./...
echo "  gofmt..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
  echo "Error: Files not formatted: $UNFORMATTED"
  exit 1
fi
echo "All checks passed."

# Show what's new since last tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -n "$LAST_TAG" ]; then
  echo ""
  echo "Changes since $LAST_TAG:"
  git log --oneline "$LAST_TAG"..HEAD
  echo ""
fi

# Tag and push
echo "Tagging $VERSION..."
git tag "$VERSION"
git push origin "$VERSION"

echo ""
echo "Released $VERSION"
echo "GoReleaser will create the GitHub Release automatically."
echo "https://github.com/henry9001/toks-go/releases/tag/$VERSION"
