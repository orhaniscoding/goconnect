# PostgreSQL Setup Guide

GoConnect supports PostgreSQL for persistent data storage. This guide explains how to set up and configure PostgreSQL for production or development use.

## Prerequisites

- PostgreSQL 15 or later
- Database user with CREATE privileges
- Network connectivity to PostgreSQL server

## Quick Start (Development)

### 1. Install PostgreSQL

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```

**macOS (Homebrew):**
```bash
brew install postgresql@15
brew services start postgresql@15
```

**Windows:**
Download and install from [postgresql.org](https://www.postgresql.org/download/windows/)

### 2. Create Database and User

```bash
# Connect as postgres user
sudo -u postgres psql

# Create database and user
CREATE DATABASE goconnect;
CREATE USER goconnect_user WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE goconnect TO goconnect_user;

# Exit psql
\q
```

### 3. Configure Environment Variables

Create a `.env` file in the `server/` directory:

```bash
# Database Connection
DB_HOST=localhost
DB_PORT=5432
DB_USER=goconnect_user
DB_PASSWORD=your_secure_password
DB_NAME=goconnect
DB_SSLMODE=disable  # Use 'require' in production

# Migration Path (optional, defaults to ./migrations)
MIGRATIONS_PATH=./migrations
```

### 4. Run Migrations

```bash
cd server
go run ./cmd/server --migrate
```

Expected output:
```
Connected to PostgreSQL: goconnect_user@localhost:5432/goconnect
Migrations applied successfully
```

### 5. Start Server

```bash
go run ./cmd/server
```

## Environment Variables

### Database Configuration

| Variable      | Description                                                 | Default     | Required |
| ------------- | ----------------------------------------------------------- | ----------- | -------- |
| `DB_HOST`     | PostgreSQL server hostname                                  | `localhost` | No       |
| `DB_PORT`     | PostgreSQL server port                                      | `5432`      | No       |
| `DB_USER`     | Database username                                           | `postgres`  | No       |
| `DB_PASSWORD` | Database password                                           | `postgres`  | No       |
| `DB_NAME`     | Database name                                               | `goconnect` | No       |
| `DB_SSLMODE`  | SSL mode (`disable`, `require`, `verify-ca`, `verify-full`) | `disable`   | No       |

### Migration Configuration

| Variable          | Description             | Default        | Required |
| ----------------- | ----------------------- | -------------- | -------- |
| `MIGRATIONS_PATH` | Path to migration files | `./migrations` | No       |

## Database Schema

GoConnect uses the following tables:

- **tenants**: Multi-tenancy support
- **users**: User accounts with Argon2id password hashing
- **networks**: VPN networks with soft delete support
- **memberships**: Network membership with roles
- **join_requests**: Network join requests (pending approval)
- **ip_allocations**: IPAM (IP Address Management)
- **idempotency_keys**: Request idempotency tracking
- **audit_events**: Audit trail (JSONB metadata)

### Key Features

- **UUID Primary Keys**: All tables use UUIDs for distributed systems
- **Foreign Key Constraints**: Referential integrity with CASCADE/SET NULL
- **Check Constraints**: Enum validation at database level
- **Indexes**: Optimized for common query patterns
- **Soft Delete**: Networks support soft delete (deleted_at column)
- **Auto Timestamps**: Triggers automatically update `updated_at`

## Migration Management

### Apply Migrations

```bash
go run ./cmd/server --migrate
```

### Rollback Migration (Manual)

GoConnect doesn't support automatic rollback via CLI. To rollback manually:

```bash
# Install golang-migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Rollback one migration
migrate -database "postgres://user:pass@localhost:5432/goconnect?sslmode=disable" \
        -path ./migrations \
        down 1
```

### Check Migration Version

```bash
# Connect to database
psql -h localhost -U goconnect_user -d goconnect

# Check version
SELECT version, dirty FROM schema_migrations;
```

## Connection Pool Settings

PostgreSQL connection pool is configured in code with these defaults:

- **MaxOpenConns**: 25 (maximum concurrent connections)
- **MaxIdleConns**: 5 (idle connections kept in pool)
- **ConnMaxLifetime**: 5 minutes
- **ConnMaxIdleTime**: 10 minutes

To optimize for your workload:

```
MaxOpenConns = (CPU cores × 2) + disk spindles
```

For example, 4 cores + 1 SSD = 9 connections.

## Production Deployment

### Security Checklist

- ✅ Use strong passwords (min 32 characters)
- ✅ Enable SSL/TLS (`DB_SSLMODE=require` or higher)
- ✅ Use dedicated database user (not `postgres`)
- ✅ Restrict network access (firewall rules)
- ✅ Enable PostgreSQL audit logging
- ✅ Regular backups (pg_dump or pg_basebackup)
- ✅ Monitor connection pool usage
- ✅ Set up replication for high availability

### SSL/TLS Configuration

```bash
# Require SSL
DB_SSLMODE=require

# Verify server certificate
DB_SSLMODE=verify-ca
DB_SSLROOTCERT=/path/to/ca.crt

# Full verification (hostname + certificate)
DB_SSLMODE=verify-full
DB_SSLROOTCERT=/path/to/ca.crt
```

### Connection String Format

Alternative to individual variables:

```bash
DATABASE_URL="postgres://user:password@host:port/database?sslmode=require"
```

Note: Current implementation uses individual env vars. Connection string support coming soon.

## Backup and Restore

### Backup

```bash
# Full backup
pg_dump -h localhost -U goconnect_user -d goconnect -F c -f goconnect_backup.dump

# Schema only
pg_dump -h localhost -U goconnect_user -d goconnect --schema-only -f schema.sql

# Data only
pg_dump -h localhost -U goconnect_user -d goconnect --data-only -f data.sql
```

### Restore

```bash
# Restore from custom dump
pg_restore -h localhost -U goconnect_user -d goconnect_new goconnect_backup.dump

# Restore from SQL file
psql -h localhost -U goconnect_user -d goconnect_new < schema.sql
```

## Monitoring

### Check Connection Count

```sql
SELECT count(*) FROM pg_stat_activity WHERE datname = 'goconnect';
```

### Check Table Sizes

```sql
SELECT
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Check Slow Queries

```sql
SELECT
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    max_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

Enable `pg_stat_statements`:
```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

## Troubleshooting

### Connection Refused

```
Error: failed to ping database: dial tcp [::1]:5432: connect: connection refused
```

**Solution**: Check PostgreSQL is running:
```bash
# Ubuntu/Debian
sudo systemctl status postgresql

# macOS
brew services list

# Windows
services.msc (check PostgreSQL service)
```

### Authentication Failed

```
Error: pq: password authentication failed for user "goconnect_user"
```

**Solution**: Check `pg_hba.conf`:
```bash
# Find config file
psql -U postgres -c "SHOW hba_file"

# Edit and add:
# TYPE  DATABASE        USER            ADDRESS         METHOD
local   goconnect       goconnect_user                  md5
host    goconnect       goconnect_user  127.0.0.1/32    md5

# Reload
sudo systemctl reload postgresql
```

### Migration Failed

```
Error: Dirty database version 1. Fix and force version.
```

**Solution**: Reset migration state:
```bash
# Connect to database
psql -h localhost -U goconnect_user -d goconnect

# Check dirty state
SELECT version, dirty FROM schema_migrations;

# Fix manually, then reset dirty flag
UPDATE schema_migrations SET dirty = false;
```

### Too Many Connections

```
Error: pq: sorry, too many clients already
```

**Solution**: Increase `max_connections` in `postgresql.conf`:
```bash
# Edit config
sudo vim /etc/postgresql/15/main/postgresql.conf

# Increase max_connections
max_connections = 100  # default is 100

# Restart PostgreSQL
sudo systemctl restart postgresql
```

## Docker Setup

Quick start with Docker Compose:

```yaml
# docker-compose.yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: goconnect
      POSTGRES_USER: goconnect_user
      POSTGRES_PASSWORD: changeme
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U goconnect_user"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

```bash
docker-compose up -d
DB_PASSWORD=changeme go run ./cmd/server --migrate
```

## Future Enhancements

- [ ] Connection string support (DATABASE_URL)
- [ ] Read replicas configuration
- [ ] Automatic migration rollback
- [ ] Migration testing framework
- [ ] Database performance profiling
- [ ] Horizontal sharding support
- [ ] Multi-region replication

## See Also

- [CONFIG_FLAGS.md](./CONFIG_FLAGS.md) - All configuration options
- [RUNBOOKS.md](./RUNBOOKS.md) - Operational runbooks
- [SECURITY.md](./SECURITY.md) - Security best practices
