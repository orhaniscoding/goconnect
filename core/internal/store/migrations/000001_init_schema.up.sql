CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    name TEXT
);

CREATE TABLE IF NOT EXISTS peers (
    id TEXT PRIMARY KEY,
    public_key TEXT NOT NULL UNIQUE,
    endpoint TEXT,
    allowed_ips TEXT
);
