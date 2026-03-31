CREATE TABLE premium_ticks (
    id           BIGSERIAL PRIMARY KEY,
    pair         VARCHAR(20) NOT NULL,
    korean_price DECIMAL(28,8) NOT NULL,
    global_price DECIMAL(28,8) NOT NULL,
    premium_pct  DECIMAL(8,4) NOT NULL,
    source_kr    VARCHAR(20) DEFAULT '',
    source_gl    VARCHAR(20) DEFAULT '',
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_premium_ticks_pair_created ON premium_ticks (pair, created_at DESC);
CREATE INDEX idx_premium_ticks_created_at ON premium_ticks (created_at DESC);
