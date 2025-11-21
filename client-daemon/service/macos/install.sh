#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Check for root
if [ "$EUID" -ne 0 ]; then 
  echo -e "${RED}Please run as root (sudo)${NC}"
  exit 1
fi

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="goconnect-daemon"
SERVICE_NAME="com.goconnect.daemon"
PLIST_FILE="com.goconnect.daemon.plist"

# Determine script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Locate binary
# Priority: 
# 1. Environment variable GOCONNECT_BIN
# 2. ../../bin/goconnect-daemon-darwin-amd64 (Cross-compiled standard name)
# 3. ../../bin/goconnect-daemon-darwin-arm64 (Apple Silicon)
# 4. ../../bin/goconnect-daemon (Local build)
# 5. Current directory

if [ -n "$GOCONNECT_BIN" ]; then
    SOURCE_BIN="$GOCONNECT_BIN"
elif [ -f "$PROJECT_ROOT/bin/goconnect-daemon-darwin-amd64" ]; then
    SOURCE_BIN="$PROJECT_ROOT/bin/goconnect-daemon-darwin-amd64"
elif [ -f "$PROJECT_ROOT/bin/goconnect-daemon-darwin-arm64" ]; then
    SOURCE_BIN="$PROJECT_ROOT/bin/goconnect-daemon-darwin-arm64"
elif [ -f "$PROJECT_ROOT/bin/goconnect-daemon" ]; then
    SOURCE_BIN="$PROJECT_ROOT/bin/goconnect-daemon"
elif [ -f "./$BINARY_NAME" ]; then
    SOURCE_BIN="./$BINARY_NAME"
else
    echo -e "${RED}Error: Could not find binary.$NC"
    echo "Please build it first with 'make build' or 'make build-all'"
    exit 1
fi

echo "Found binary at: $SOURCE_BIN"

# Unload existing service
if launchctl list | grep -q "$SERVICE_NAME"; then
    echo "Unloading existing service..."
    launchctl unload "/Library/LaunchDaemons/$PLIST_FILE" || true
fi

# Install Binary
echo "Installing binary to $INSTALL_DIR/$BINARY_NAME..."
mkdir -p "$INSTALL_DIR"
cp "$SOURCE_BIN" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Install Plist
echo "Installing launchd plist..."
cp "$SCRIPT_DIR/$PLIST_FILE" "/Library/LaunchDaemons/$PLIST_FILE"
chmod 644 "/Library/LaunchDaemons/$PLIST_FILE"
chown root:wheel "/Library/LaunchDaemons/$PLIST_FILE"

# Load Service
echo "Loading service..."
launchctl load "/Library/LaunchDaemons/$PLIST_FILE"

echo -e "${GREEN}✅ GoConnect Daemon installed and started successfully!${NC}"
echo "Check status with: sudo launchctl list | grep $SERVICE_NAME"
