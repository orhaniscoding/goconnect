#!/bin/bash
set -e

# Check for root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root (sudo)"
  exit 1
fi

BINARY="goconnect-daemon"
INSTALL_DIR="/usr/local/bin"
PLIST_FILE="com.goconnect.daemon.plist"
LAUNCHDAEMONS="/Library/LaunchDaemons"

# Check if binary exists in current directory
if [ ! -f "./$BINARY" ]; then
  echo "Error: $BINARY not found in current directory"
  exit 1
fi

# Unload existing service
if launchctl list | grep -q com.goconnect.daemon; then
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

echo "✅ GoConnect Daemon installed successfully!"
echo ""
echo "Next steps:"
echo "1. Create config: /etc/goconnect/config.yaml"
echo "2. Service will start automatically"
echo "3. Check status: sudo launchctl list | grep goconnect"
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
