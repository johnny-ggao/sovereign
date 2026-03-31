CREATE TABLE wallets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency     VARCHAR(10) NOT NULL,
    available    DECIMAL(28,18) NOT NULL DEFAULT 0,
    in_operation DECIMAL(28,18) NOT NULL DEFAULT 0,
    frozen       DECIMAL(28,18) NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, currency)
);

CREATE INDEX idx_wallets_user_id ON wallets (user_id);
