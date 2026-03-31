CREATE TABLE settlements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investment_id   UUID NOT NULL REFERENCES investments(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period          VARCHAR(7) NOT NULL,
    gross_return    DECIMAL(28,18) NOT NULL,
    performance_fee DECIMAL(28,18) NOT NULL,
    fee_rate        DECIMAL(5,4) NOT NULL DEFAULT 0.5,
    net_return      DECIMAL(28,18) NOT NULL,
    trade_count     INT NOT NULL DEFAULT 0,
    avg_premium_pct DECIMAL(8,4) DEFAULT 0,
    report_url      VARCHAR(500) DEFAULT '',
    settled_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(investment_id, period)
);

CREATE INDEX idx_settlements_user_id ON settlements (user_id);
CREATE INDEX idx_settlements_period ON settlements (period);
