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

# Create configuration directory and example config
CONFIG_DIR="/etc/goconnect"
CONFIG_FILE="$CONFIG_DIR/.env"
EXAMPLE_CONFIG="./config.example.env"

if [ ! -d "$CONFIG_DIR" ]; then
  mkdir -p "$CONFIG_DIR"
  echo "Created config directory: $CONFIG_DIR"
fi

if [ ! -f "$CONFIG_FILE" ]; then
  if [ -f "$EXAMPLE_CONFIG" ]; then
    cp "$EXAMPLE_CONFIG" "$CONFIG_FILE"
    chmod 600 "$CONFIG_FILE"
    echo "Created example config: $CONFIG_FILE"
    echo "⚠️  IMPORTANT: Edit this file with database and WireGuard settings!"
  else
    # Create minimal config if example doesn't exist
    cat > "$CONFIG_FILE" <<EOF
# GoConnect Server Configuration
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goconnect
DB_USER=goconnect
DB_PASSWORD=your_secure_password
JWT_SECRET=your_jwt_secret_min_32_chars
WG_SERVER_ENDPOINT=vpn.example.com:51820
WG_SERVER_PUBKEY=your_wireguard_public_key_44_chars
EOF
    chmod 600 "$CONFIG_FILE"
    echo "Created minimal config: $CONFIG_FILE"
    echo "⚠️  IMPORTANT: Edit this file before starting!"
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

echo "✅ GoConnect Server installed successfully!"
echo ""
echo "REQUIRED: Configure before starting:"
echo "1. Edit: $CONFIG_FILE"
echo "2. Set database credentials (DB_*)"
echo "3. Set JWT_SECRET (min 32 chars)"
echo "4. Set WireGuard settings (WG_*)"
echo "5. Run migrations: $INSTALL_DIR/$BINARY -migrate"
echo "6. Start service: sudo systemctl start goconnect-server"
echo "7. Enable on boot: sudo systemctl enable goconnect-server"
echo ""
echo "See config.example.env for all available options."
