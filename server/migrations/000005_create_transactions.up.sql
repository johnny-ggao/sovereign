CREATE TABLE transactions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type         VARCHAR(20) NOT NULL,
    currency     VARCHAR(10) NOT NULL,
    network      VARCHAR(20) DEFAULT '',
    amount       DECIMAL(28,18) NOT NULL,
    fee          DECIMAL(28,18) NOT NULL DEFAULT 0,
    address      VARCHAR(255) DEFAULT '',
    tx_hash      VARCHAR(255) DEFAULT '',
    status       VARCHAR(20) DEFAULT 'pending',
    external_id  VARCHAR(255) DEFAULT '',
    confirmed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON transactions (user_id);
CREATE INDEX idx_transactions_status ON transactions (status);
CREATE INDEX idx_transactions_type ON transactions (type);
CREATE INDEX idx_transactions_external_id ON transactions (external_id);
CREATE INDEX idx_transactions_created_at ON transactions (created_at DESC);
