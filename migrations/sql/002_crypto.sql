CREATE TABLE IF NOT EXISTS crypto (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    initial TEXT NOT NULL,
    name TEXT NOT NULL,
    current_value DECIMAL(19, 8) NOT NULL,
    previous_value DECIMAL(19, 8),
    percent DECIMAL(10, 3) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_initial BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(name, created_at)
);

CREATE INDEX IF NOT EXISTS idx_crypto_name_initial_created_at ON crypto (name, initial, created_at DESC);