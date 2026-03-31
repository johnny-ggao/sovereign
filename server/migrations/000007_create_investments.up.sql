CREATE TABLE investments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount          DECIMAL(28,18) NOT NULL,
    currency        VARCHAR(10) NOT NULL DEFAULT 'USDT',
    status          VARCHAR(20) DEFAULT 'active',
    total_return    DECIMAL(28,18) NOT NULL DEFAULT 0,
    performance_fee DECIMAL(28,18) NOT NULL DEFAULT 0,
    net_return      DECIMAL(28,18) NOT NULL DEFAULT 0,
    start_date      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_date        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_investments_user_id ON investments (user_id);
CREATE INDEX idx_investments_status ON investments (status);
