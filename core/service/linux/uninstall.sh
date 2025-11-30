#!/bin/bash
set -e

# Check for root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root (sudo)"
  exit 1
fi

BINARY="goconnect-server"
INSTALL_DIR="/usr/local/bin"
SYSTEMD_DIR="/etc/systemd/system"

# Stop and disable
if systemctl is-active --quiet goconnect-server; then
  echo "Stopping service..."
  systemctl stop goconnect-server
fi

if systemctl is-enabled --quiet goconnect-server 2>/dev/null; then
  echo "Disabling service..."
  systemctl disable goconnect-server
fi

# Remove service file
if [ -f "$SYSTEMD_DIR/goconnect-server.service" ]; then
  echo "Removing service file..."
  rm "$SYSTEMD_DIR/goconnect-server.service"
  systemctl daemon-reload
fi

# Remove binary
if [ -f "$INSTALL_DIR/$BINARY" ]; then
  echo "Removing binary..."
  rm "$INSTALL_DIR/$BINARY"
fi

echo "✅ GoConnect Server uninstalled successfully!"
echo ""
echo "Note: Config files in /etc/goconnect were preserved."
echo "To remove: sudo rm -rf /etc/goconnect"
