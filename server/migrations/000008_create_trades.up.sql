CREATE TABLE trades (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investment_id  UUID NOT NULL REFERENCES investments(id) ON DELETE CASCADE,
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pair           VARCHAR(20) NOT NULL,
    buy_exchange   VARCHAR(20) NOT NULL,
    sell_exchange   VARCHAR(20) NOT NULL,
    buy_price      DECIMAL(28,8) NOT NULL,
    sell_price     DECIMAL(28,8) NOT NULL,
    amount         DECIMAL(28,18) NOT NULL,
    premium_pct    DECIMAL(8,4) NOT NULL,
    pnl            DECIMAL(28,18) NOT NULL,
    fee            DECIMAL(28,18) NOT NULL DEFAULT 0,
    executed_at    TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_trades_investment_id ON trades (investment_id);
CREATE INDEX idx_trades_user_id ON trades (user_id);
CREATE INDEX idx_trades_executed_at ON trades (executed_at DESC);
CREATE INDEX idx_trades_pair ON trades (pair);
