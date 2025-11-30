# GoConnect Server - Linux Installation

## Quick Install

```bash
tar -xzf goconnect-server_*_linux_*.tar.gz
cd goconnect-server_*_linux_*
chmod +x install.sh
sudo ./install.sh
```

The installer will:
1. Install binary to `/usr/local/bin/goconnect-server`
2. Create systemd service `goconnect-server.service`
3. Create config directory `/etc/goconnect/`
4. Create example config `/etc/goconnect/.env`

## Configuration

**REQUIRED**: Edit the configuration file before starting.

```bash
sudo nano /etc/goconnect/.env
```

### Essential Settings

```bash
# Database (PostgreSQL required)
DB_HOST=localhost
DB_NAME=goconnect
DB_USER=goconnect
DB_PASSWORD=your_secure_password

# JWT Secret (min 32 chars)
JWT_SECRET=your_jwt_secret_min_32_chars

# WireGuard
WG_SERVER_ENDPOINT=vpn.example.com:51820
WG_SERVER_PUBKEY=your_44_char_public_key
WG_PRIVATE_KEY=your_private_key
```

### Generate Keys

```bash
# JWT Secret
openssl rand -base64 32

# WireGuard Keys
wg genkey  # Private
echo "YOUR_PRIVATE_KEY" | wg pubkey  # Public
```

See `config.example.env` for all available options.

## Database Setup

```bash
# Install PostgreSQL
sudo apt install postgresql

# Create database
sudo -u postgres psql
CREATE DATABASE goconnect;
CREATE USER goconnect WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE goconnect TO goconnect;
\q
```

## Run Migrations

```bash
sudo /usr/local/bin/goconnect-server -migrate
```

## Service Management

```bash
# Start
sudo systemctl start goconnect-server

# Stop
sudo systemctl stop goconnect-server

# Status
sudo systemctl status goconnect-server

# Enable on boot
sudo systemctl enable goconnect-server

# Logs
sudo journalctl -u goconnect-server -f
```

## Uninstall

```bash
sudo ./uninstall.sh
```

## Troubleshooting

See `/docs/CONFIGURATION.md` for complete guide.

### Quick Checks

```bash
# Check config
ls -l /etc/goconnect/.env

# Test database
psql -h localhost -U goconnect -d goconnect

# Check logs
sudo journalctl -u goconnect-server -n 50
```

## Requirements

- Linux kernel 5.6+
- PostgreSQL 12+
- WireGuard: `sudo apt install wireguard-tools`
- systemd
