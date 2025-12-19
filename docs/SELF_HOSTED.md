# 🏠 GoConnect Self-Hosted Quick Start

This guide helps you deploy your own GoConnect Server instance using Docker Compose.

## Prerequisites

*   Docker Engine & Docker Compose
*   A public domain name (e.g., `vpn.example.com`)
*   Ports `80`, `443`, and `51820/udp` open

## 🚀 Speed Run

1.  **Download the Compose file**:
    ```bash
    curl -o docker-compose.yml https://raw.githubusercontent.com/orhaniscoding/goconnect/main/deploy/docker-compose.yml
    ```

2.  **Generate Secrets**:
    ```bash
    # Generate a strong JWT secret
    openssl rand -base64 32
    ```

3.  **Configure Environment**:
    Create a `.env` file:
    ```ini
    SERVER_HOST=vpn.example.com
    JWT_SECRET=your_generated_secret_here
    POSTGRES_PASSWORD=secure_db_password
    ```

4.  **Start the Stack**:
    ```bash
    docker compose up -d
    ```

5.  **Create Admin User**:
    ```bash
    docker compose exec server goconnect-server users create --email admin@example.com --password "AdminPass123!" --role admin
    ```

## 📦 What's Included?

*   **GoConnect Server**: The API and control plane.
*   **PostgreSQL**: Users, Networks, and Config storage.
*   **Redis**: Real-time events and caching.
*   **Caddy (Optional)**: Automatic HTTPS reverse proxy (add to compose if needed).

## 🔧 Troubleshooting

**Logs:**
```bash
docker compose logs -f server
```

**Database Connection:**
If the server can't connect to DB, ensure `goconnect` network is created:
```bash
docker network ls
```
