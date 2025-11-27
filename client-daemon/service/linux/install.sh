#!/bin/bash
set -e

# Check for root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root (sudo)"
  exit 1
fi

BINARY="goconnect-daemon"
INSTALL_DIR="/usr/local/bin"
SERVICE_FILE="goconnect-daemon.service"
SYSTEMD_DIR="/etc/systemd/system"

# Check if binary exists in current directory
if [ ! -f "./$BINARY" ]; then
  echo "Error: $BINARY not found in current directory"
  exit 1
fi

# Stop existing service
if systemctl is-active --quiet goconnect-daemon; then
  echo "Stopping existing service..."
  systemctl stop goconnect-daemon
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

echo "✅ GoConnect Daemon installed successfully!"
echo ""
echo "Next steps:"
echo "1. Create config: /etc/goconnect/config.yaml"
echo "2. Start service: sudo systemctl start goconnect-daemon"
echo "3. Enable on boot: sudo systemctl enable goconnect-daemon"
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
