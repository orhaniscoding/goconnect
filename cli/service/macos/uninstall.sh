#!/bin/bash
set -e

# Check for root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root (sudo)"
  exit 1
fi

BINARY="goconnect"
INSTALL_DIR="/usr/local/bin"
LAUNCHDAEMONS="/Library/LaunchDaemons"
PLIST="com.goconnect.plist"
SERVICE_NAME="com.goconnect"

# Unload service
if launchctl list | grep -q "$SERVICE_NAME"; then
  echo "Stopping service..."
  launchctl unload "$LAUNCHDAEMONS/$PLIST" 2>/dev/null || true
fi

# Remove plist
if [ -f "$LAUNCHDAEMONS/$PLIST" ]; then
  echo "Removing LaunchDaemon..."
  rm "$LAUNCHDAEMONS/$PLIST"
fi

# Remove binary
if [ -f "$INSTALL_DIR/$BINARY" ]; then
  echo "Removing binary..."
  rm "$INSTALL_DIR/$BINARY"
fi

echo "✅ GoConnect uninstalled successfully!"

