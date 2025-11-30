-- Rollback initial schema migration
-- Version: 1.0
-- Author: orhaniscoding
-- Date: 2025-10-29

-- Drop triggers
DROP TRIGGER IF EXISTS update_memberships_updated_at ON memberships;
DROP TRIGGER IF EXISTS update_networks_updated_at ON networks;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS audit_events;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS ip_allocations;
DROP TABLE IF EXISTS join_requests;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS networks;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
