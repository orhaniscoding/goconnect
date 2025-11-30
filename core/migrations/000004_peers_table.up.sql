-- Create peers table
CREATE TABLE IF NOT EXISTS peers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    network_id UUID NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    public_key VARCHAR(44) NOT NULL,
    preshared_key VARCHAR(44),
    endpoint VARCHAR(255),
    allowed_ips TEXT[] NOT NULL,
    persistent_keepalive INTEGER DEFAULT 0,
    last_handshake TIMESTAMP WITH TIME ZONE,
    rx_bytes BIGINT DEFAULT 0,
    tx_bytes BIGINT DEFAULT 0,
    active BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    disabled_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT peers_persistent_keepalive_check CHECK (persistent_keepalive >= 0 AND persistent_keepalive <= 65535)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_peers_network_id ON peers(network_id) WHERE disabled_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_peers_device_id ON peers(device_id) WHERE disabled_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_peers_tenant_id ON peers(tenant_id) WHERE disabled_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_peers_public_key ON peers(public_key) WHERE disabled_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_peers_active ON peers(network_id, active) WHERE disabled_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_peers_last_handshake ON peers(last_handshake DESC NULLS LAST) WHERE disabled_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_peers_network_device_unique ON peers(network_id, device_id) WHERE disabled_at IS NULL;

-- Trigger for updated_at
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
