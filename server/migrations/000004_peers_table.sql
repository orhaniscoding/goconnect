-- Migration: Create peers table for WireGuard peer management
-- Description: Stores WireGuard peer information including keys, endpoints, and statistics

-- Create peers table
CREATE TABLE IF NOT EXISTS peers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    network_id UUID NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- WireGuard configuration
    public_key VARCHAR(44) NOT NULL, -- Base64 encoded 32-byte key
    preshared_key VARCHAR(44), -- Optional preshared key for additional security
    endpoint VARCHAR(255), -- Last known endpoint (IP:Port)
    allowed_ips TEXT[] NOT NULL, -- CIDR blocks this peer can route
    persistent_keepalive INTEGER DEFAULT 0, -- Keepalive interval in seconds (0 = disabled)
    
    -- Statistics and status
    last_handshake TIMESTAMP WITH TIME ZONE, -- Last successful WireGuard handshake
    rx_bytes BIGINT DEFAULT 0, -- Received bytes
    tx_bytes BIGINT DEFAULT 0, -- Transmitted bytes
    active BOOLEAN DEFAULT false, -- Whether peer is currently active
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    disabled_at TIMESTAMP WITH TIME ZONE, -- Soft delete timestamp
    
    -- Constraints
    CONSTRAINT peers_persistent_keepalive_check CHECK (persistent_keepalive >= 0 AND persistent_keepalive <= 65535),
    CONSTRAINT peers_public_key_length CHECK (length(public_key) = 44),
    CONSTRAINT peers_preshared_key_length CHECK (preshared_key IS NULL OR length(preshared_key) = 44)
);

-- Create indexes for performance
CREATE INDEX idx_peers_network_id ON peers(network_id) WHERE disabled_at IS NULL;
CREATE INDEX idx_peers_device_id ON peers(device_id) WHERE disabled_at IS NULL;
CREATE INDEX idx_peers_tenant_id ON peers(tenant_id) WHERE disabled_at IS NULL;
CREATE INDEX idx_peers_public_key ON peers(public_key) WHERE disabled_at IS NULL;
CREATE INDEX idx_peers_active ON peers(network_id, active) WHERE disabled_at IS NULL;
CREATE INDEX idx_peers_last_handshake ON peers(last_handshake DESC NULLS LAST) WHERE disabled_at IS NULL;

-- Create unique constraint for active peers
-- A device can only have one active peer per network
CREATE UNIQUE INDEX idx_peers_network_device_unique ON peers(network_id, device_id) WHERE disabled_at IS NULL;

-- Note: We do NOT enforce public key uniqueness globally because the same device
-- (with same public key) can participate in multiple networks. The unique constraint
-- is only on (network_id, device_id) combination to prevent duplicate peers.

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_peers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_peers_updated_at
    BEFORE UPDATE ON peers
    FOR EACH ROW
    EXECUTE FUNCTION update_peers_updated_at();

-- Add comments for documentation
COMMENT ON TABLE peers IS 'WireGuard peers representing device connections in networks';
COMMENT ON COLUMN peers.public_key IS 'WireGuard public key (base64 encoded Curve25519 key)';
COMMENT ON COLUMN peers.preshared_key IS 'Optional preshared key for post-quantum security';
COMMENT ON COLUMN peers.endpoint IS 'Last known peer endpoint (IP:Port)';
COMMENT ON COLUMN peers.allowed_ips IS 'CIDR blocks this peer is allowed to route';
COMMENT ON COLUMN peers.persistent_keepalive IS 'Keepalive interval in seconds (0 = disabled, recommended: 25)';
COMMENT ON COLUMN peers.last_handshake IS 'Timestamp of last successful WireGuard handshake';
COMMENT ON COLUMN peers.rx_bytes IS 'Total bytes received from this peer';
COMMENT ON COLUMN peers.tx_bytes IS 'Total bytes transmitted to this peer';
COMMENT ON COLUMN peers.active IS 'Whether peer is currently active (recent handshake)';
