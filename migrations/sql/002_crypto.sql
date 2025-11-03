CREATE TABLE IF NOT EXISTS crypto (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    initial TEXT NOT NULL,
    name TEXT NOT NULL,
    current_value DECIMAL(19, 8) NOT NULL,
    previous_value DECIMAL(19, 8),
    procent DECIMAL(10, 4) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_crypto_name_initial_created_at ON crypto (name, initial, created_at DESC);