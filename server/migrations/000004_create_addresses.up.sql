CREATE TABLE deposit_addresses (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency   VARCHAR(10) NOT NULL,
    network    VARCHAR(20) NOT NULL,
    address    VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_deposit_addresses_user_id ON deposit_addresses (user_id);
CREATE UNIQUE INDEX idx_deposit_addresses_unique ON deposit_addresses (user_id, currency, network);

CREATE TABLE withdraw_addresses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency        VARCHAR(10) NOT NULL,
    network         VARCHAR(20) NOT NULL,
    address         VARCHAR(255) NOT NULL,
    label           VARCHAR(100) DEFAULT '',
    cooldown_until  TIMESTAMPTZ NOT NULL,
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_withdraw_addresses_user_id ON withdraw_addresses (user_id);
