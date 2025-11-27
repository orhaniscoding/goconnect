#!/bin/bash
set -e

# Check for root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root (sudo)"
  exit 1
fi

BINARY="goconnect-server"
INSTALL_DIR="/usr/local/bin"
PLIST_FILE="com.goconnect.server.plist"
LAUNCHDAEMONS="/Library/LaunchDaemons"

# Check binary
if [ ! -f "./$BINARY" ]; then
  echo "Error: $BINARY not found in current directory"
  exit 1
fi

# Unload existing service
if launchctl list | grep -q com.goconnect.server; then
  echo "Stopping existing service..."
  launchctl unload "$LAUNCHDAEMONS/$PLIST_FILE" 2>/dev/null || true
fi

# Copy binary
echo "Installing binary to $INSTALL_DIR..."
cp "./$BINARY" "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"

# Install LaunchDaemon
if [ -f "./$PLIST_FILE" ]; then
  echo "Installing LaunchDaemon..."
  cp "./$PLIST_FILE" "$LAUNCHDAEMONS/$PLIST_FILE"
  chmod 644 "$LAUNCHDAEMONS/$PLIST_FILE"
  launchctl load "$LAUNCHDAEMONS/$PLIST_FILE"
fi

echo "✅ GoConnect Server installed successfully!"
echo ""
echo "Next steps:"
echo "1. Create config: /etc/goconnect/config.yaml"
echo "2. Service will start automatically"
echo "3. Check status: sudo launchctl list | grep goconnect"
