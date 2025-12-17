#!/bin/bash
set -e

# GoConnect Release Script

if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 v0.1.0"
  exit 1
fi

VERSION=$1

# 1. Check for clean working tree
if [ -n "$(git status --porcelain)" ]; then
  echo "❌ Error: Working tree is not clean. Please commit or stash changes."
  exit 1
fi

echo "🚀 preparing release $VERSION..."

# 2. Run Tests
echo "🧪 Running Core tests..."
(cd core && go test -short ./...)

echo "🧪 Running CLI tests (pure Go)..."
# Ensure we test with CGO_ENABLED=0 to match release build
(cd cli && CGO_ENABLED=0 go test -short ./...)

# 3. Create Tag
echo "🏷️ Tagging version $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

# 4. Push
echo "⬆️ Pushing tag to origin..."
git push origin "$VERSION"

echo "✅ Release $VERSION triggered!"
echo "Check progress at: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\(.*\).git/\1/')/actions"
