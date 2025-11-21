#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Check for root
if [ "$EUID" -ne 0 ]; then 
  echo -e "${RED}Please run as root${NC}"
  exit 1
fi

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="goconnect-daemon"
SERVICE_NAME="goconnect-daemon"
SERVICE_FILE="goconnect-daemon.service"

# Determine script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Locate binary
# Priority: 
# 1. Environment variable GOCONNECT_BIN
# 2. ../../bin/goconnect-daemon-linux-amd64 (Cross-compiled standard name)
# 3. ../../bin/goconnect-daemon (Local build)
# 4. Current directory

if [ -n "$GOCONNECT_BIN" ]; then
    SOURCE_BIN="$GOCONNECT_BIN"
elif [ -f "$PROJECT_ROOT/bin/goconnect-daemon-linux-amd64" ]; then
    SOURCE_BIN="$PROJECT_ROOT/bin/goconnect-daemon-linux-amd64"
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

# Stop existing service
if systemctl is-active --quiet $SERVICE_NAME; then
    echo "Stopping existing service..."
    systemctl stop $SERVICE_NAME
fi

# Install Binary
echo "Installing binary to $INSTALL_DIR/$BINARY_NAME..."
cp "$SOURCE_BIN" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Install Service
echo "Installing systemd service..."
cp "$SCRIPT_DIR/$SERVICE_FILE" "/etc/systemd/system/$SERVICE_FILE"

# Reload and Enable
echo "Reloading systemd..."
systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME

echo -e "${GREEN}✅ GoConnect Daemon installed and started successfully!${NC}"
echo "Check status with: systemctl status $SERVICE_NAME"
