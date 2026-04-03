CREATE TABLE user_trades (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    investment_id  UUID NOT NULL REFERENCES investments(id) ON DELETE CASCADE,
    trade_id       UUID REFERENCES trades(id) ON DELETE SET NULL,
    settlement_id  UUID REFERENCES settlements(id) ON DELETE SET NULL,
    pair           VARCHAR(20) NOT NULL,
    buy_exchange   VARCHAR(20) NOT NULL,
    sell_exchange  VARCHAR(20) NOT NULL,
    buy_price      DECIMAL(38,8) NOT NULL,
    sell_price     DECIMAL(38,8) NOT NULL,
    amount         DECIMAL(28,18) NOT NULL,
    premium_pct    DECIMAL(8,4) NOT NULL,
    pnl            DECIMAL(28,18) NOT NULL,
    fee            DECIMAL(28,18) NOT NULL DEFAULT 0,
    ratio          DECIMAL(10,8) NOT NULL,
    executed_at    TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_user_trades_user_id ON user_trades (user_id);
CREATE INDEX idx_user_trades_investment_id ON user_trades (investment_id);
CREATE INDEX idx_user_trades_executed_at ON user_trades (executed_at DESC);
