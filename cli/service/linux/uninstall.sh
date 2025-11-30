#!/bin/bash
set -e

# Check for root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root (sudo)"
  exit 1
fi

BINARY="goconnect-daemon"
INSTALL_DIR="/usr/local/bin"
SYSTEMD_DIR="/etc/systemd/system"

# Stop and disable service
if systemctl is-active --quiet goconnect-daemon; then
  echo "Stopping service..."
  systemctl stop goconnect-daemon
fi

if systemctl is-enabled --quiet goconnect-daemon 2>/dev/null; then
  echo "Disabling service..."
  systemctl disable goconnect-daemon
fi

# Remove service file
if [ -f "$SYSTEMD_DIR/goconnect-daemon.service" ]; then
  echo "Removing service file..."
  rm "$SYSTEMD_DIR/goconnect-daemon.service"
  systemctl daemon-reload
fi

# Remove binary
if [ -f "$INSTALL_DIR/$BINARY" ]; then
  echo "Removing binary..."
  rm "$INSTALL_DIR/$BINARY"
fi

echo "✅ GoConnect Daemon uninstalled successfully!"
echo ""
echo "Note: Config files in /etc/goconnect were preserved."
echo "To remove: sudo rm -rf /etc/goconnect"
