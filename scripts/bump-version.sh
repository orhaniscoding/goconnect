#!/usr/bin/env bash
#
# Version bump script for GoConnect
# Usage: ./scripts/bump-version.sh <new-version>
# Example: ./scripts/bump-version.sh 3.1.0
#
# This script updates version numbers across all project files:
# - desktop/package.json
# - desktop/src-tauri/tauri.conf.json
# - desktop/src-tauri/Cargo.toml
#
# After running this script, commit the changes and create a tag:
#   git add -A
#   git commit -m "chore: bump version to v$NEW_VERSION"
#   git tag v$NEW_VERSION
#   git push origin main --tags
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Function to print colored output
print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }

# Check arguments
if [ -z "$1" ]; then
    print_error "Usage: $0 <new-version>"
    print_info "Example: $0 3.1.0"
    exit 1
fi

NEW_VERSION="$1"

# Validate version format (semver)
if ! [[ "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
    print_error "Invalid version format: $NEW_VERSION"
    print_info "Version must be in semver format: X.Y.Z or X.Y.Z-prerelease"
    exit 1
fi

print_info "Bumping version to: $NEW_VERSION"
echo ""

# Get current versions
CURRENT_PACKAGE_VERSION=$(jq -r .version "$PROJECT_ROOT/desktop/package.json")
CURRENT_TAURI_VERSION=$(jq -r .version "$PROJECT_ROOT/desktop/src-tauri/tauri.conf.json")
CURRENT_CARGO_VERSION=$(grep '^version = ' "$PROJECT_ROOT/desktop/src-tauri/Cargo.toml" | head -1 | sed 's/version = "\(.*\)"/\1/')

print_info "Current versions:"
echo "  - package.json:      $CURRENT_PACKAGE_VERSION"
echo "  - tauri.conf.json:   $CURRENT_TAURI_VERSION"
echo "  - Cargo.toml:        $CURRENT_CARGO_VERSION"
echo ""

# Update package.json
print_info "Updating desktop/package.json..."
jq ".version = \"$NEW_VERSION\"" "$PROJECT_ROOT/desktop/package.json" > "$PROJECT_ROOT/desktop/package.json.tmp"
mv "$PROJECT_ROOT/desktop/package.json.tmp" "$PROJECT_ROOT/desktop/package.json"
print_success "Updated package.json"

# Update tauri.conf.json
print_info "Updating desktop/src-tauri/tauri.conf.json..."
jq ".version = \"$NEW_VERSION\"" "$PROJECT_ROOT/desktop/src-tauri/tauri.conf.json" > "$PROJECT_ROOT/desktop/src-tauri/tauri.conf.json.tmp"
mv "$PROJECT_ROOT/desktop/src-tauri/tauri.conf.json.tmp" "$PROJECT_ROOT/desktop/src-tauri/tauri.conf.json"
print_success "Updated tauri.conf.json"

# Update Cargo.toml
print_info "Updating desktop/src-tauri/Cargo.toml..."
sed -i "s/^version = \".*\"/version = \"$NEW_VERSION\"/" "$PROJECT_ROOT/desktop/src-tauri/Cargo.toml"
print_success "Updated Cargo.toml"

echo ""
print_success "All versions updated to: $NEW_VERSION"
echo ""

# Verify updates
print_info "Verifying updates..."
UPDATED_PACKAGE=$(jq -r .version "$PROJECT_ROOT/desktop/package.json")
UPDATED_TAURI=$(jq -r .version "$PROJECT_ROOT/desktop/src-tauri/tauri.conf.json")
UPDATED_CARGO=$(grep '^version = ' "$PROJECT_ROOT/desktop/src-tauri/Cargo.toml" | head -1 | sed 's/version = "\(.*\)"/\1/')

if [ "$UPDATED_PACKAGE" = "$NEW_VERSION" ] && [ "$UPDATED_TAURI" = "$NEW_VERSION" ] && [ "$UPDATED_CARGO" = "$NEW_VERSION" ]; then
    print_success "All versions verified!"
else
    print_error "Version mismatch detected!"
    echo "  - package.json:      $UPDATED_PACKAGE"
    echo "  - tauri.conf.json:   $UPDATED_TAURI"
    echo "  - Cargo.toml:        $UPDATED_CARGO"
    exit 1
fi

echo ""
print_info "Next steps:"
echo "  1. Review changes: git diff"
echo "  2. Commit: git add -A && git commit -m 'chore: bump version to v$NEW_VERSION'"
echo "  3. Tag: git tag v$NEW_VERSION"
echo "  4. Push: git push origin main --tags"
echo ""
print_success "Done!"
