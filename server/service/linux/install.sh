#!/bin/bash
set -e

# Check for root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root (sudo)"
  exit 1
fi

BINARY="goconnect-server"
INSTALL_DIR="/usr/local/bin"
SERVICE_FILE="goconnect-server.service"
SYSTEMD_DIR="/etc/systemd/system"

# Check binary
if [ ! -f "./$BINARY" ]; then
  echo "Error: $BINARY not found in current directory"
  exit 1
fi

# Stop existing service
if systemctl is-active --quiet goconnect-server; then
  echo "Stopping existing service..."
  systemctl stop goconnect-server
fi

# Copy binary
echo "Installing binary to $INSTALL_DIR..."
cp "./$BINARY" "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"

# Install systemd service
if [ -f "./$SERVICE_FILE" ]; then
  echo "Installing systemd service..."
  cp "./$SERVICE_FILE" "$SYSTEMD_DIR/$SERVICE_FILE"
  systemctl daemon-reload
fi

echo "✅ GoConnect Server installed successfully!"
echo ""
echo "Next steps:"
echo "1. Create config: /etc/goconnect/config.yaml"
echo "2. Start service: sudo systemctl start goconnect-server"
echo "3. Enable on boot: sudo systemctl enable goconnect-server"
