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

# Create configuration directory and example config
CONFIG_DIR="/etc/goconnect"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
EXAMPLE_CONFIG="./config.example.yaml"

if [ ! -d "$CONFIG_DIR" ]; then
  mkdir -p "$CONFIG_DIR"
  echo "Created config directory: $CONFIG_DIR"
fi

if [ ! -f "$CONFIG_FILE" ]; then
  if [ -f "$EXAMPLE_CONFIG" ]; then
    cp "$EXAMPLE_CONFIG" "$CONFIG_FILE"
    chmod 600 "$CONFIG_FILE"
    echo "Created example config: $CONFIG_FILE"
    echo "⚠️  IMPORTANT: Edit this file with your server URL before starting the service!"
  else
    # Create minimal config if example doesn't exist
    cat > "$CONFIG_FILE" <<EOF
# GoConnect Daemon Configuration
# REQUIRED: Set your server URL
server_url: "https://vpn.example.com:8080"

# Optional settings (defaults shown)
local_port: 12345
log_level: "info"
interface_name: "wg0"
EOF
    chmod 600 "$CONFIG_FILE"
    echo "Created minimal config: $CONFIG_FILE"
    echo "⚠️  IMPORTANT: Edit this file with your server URL before starting the service!"
  fi
else
  echo "Config file already exists: $CONFIG_FILE"
fi

# Install systemd service
if [ -f "./$SERVICE_FILE" ]; then
  echo "Installing systemd service..."
  cp "./$SERVICE_FILE" "$SYSTEMD_DIR/$SERVICE_FILE"
  systemctl daemon-reload
fi

echo "✅ GoConnect Daemon installed successfully!"
echo ""
echo "REQUIRED: Configure before starting:"
echo "1. Edit: $CONFIG_FILE"
echo "2. Set your server_url"
echo "3. Start service: sudo systemctl start goconnect-daemon"
echo "4. Enable on boot: sudo systemctl enable goconnect-daemon"
echo ""
echo "See config.example.yaml for all available options."
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
