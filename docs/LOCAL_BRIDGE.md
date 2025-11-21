# Local Client Daemon Bridge

The GoConnect Client Daemon runs a local HTTP server to allow the Web UI to interact with the WireGuard interface and manage device registration.

## Configuration

- **Port:** Fixed at `12345` (TCP)
- **Bind Address:** `127.0.0.1` (Localhost only)
- **CORS:** Enabled for `*` (Development) / Restricted origins (Production)

## API Endpoints

### 1. Get Status
Returns the current status of the daemon, WireGuard interface, and device registration.

- **URL:** `GET http://127.0.0.1:12345/status`
- **Response:**
  ```json
  {
    "running": true,
    "version": "0.1.0",
    "wg": {
      "active": true,
      "public_key": "..."
    },
    "device": {
      "registered": true,
      "device_id": "dev_...",
      "public_key": "..."
    }
  }
  ```

### 2. Register Device
Registers the local device with the GoConnect Server using a user access token.

- **URL:** `POST http://127.0.0.1:12345/register`
- **Body:**
  ```json
  {
    "token": "user_access_token_here"
  }
  ```
- **Response:**
  ```json
  {
    "status": "success",
    "device_id": "dev_..."
  }
  ```

### 3. Connect VPN
Enables the VPN connection and starts syncing configuration.

- **URL:** `POST http://127.0.0.1:12345/connect`
- **Response:**
  ```json
  {
    "status": "connected"
  }
  ```

### 4. Disconnect VPN
Disables the VPN connection and clears the WireGuard interface.

- **URL:** `POST http://127.0.0.1:12345/disconnect`
- **Response:**
  ```json
  {
    "status": "disconnected"
  }
  ```

## Security

- The bridge only listens on the loopback interface (`127.0.0.1`).
- Sensitive operations (like registration) require a valid user token from the Web UI.


