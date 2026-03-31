CREATE TABLE notification_prefs (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email_trade        BOOLEAN DEFAULT true,
    email_deposit      BOOLEAN DEFAULT true,
    email_withdraw     BOOLEAN DEFAULT true,
    email_settlement   BOOLEAN DEFAULT true,
    push_premium_alert BOOLEAN DEFAULT false,
    push_trade         BOOLEAN DEFAULT true,
    push_deposit       BOOLEAN DEFAULT true,
    push_withdraw      BOOLEAN DEFAULT true,
    premium_threshold  DECIMAL(5,2) DEFAULT 3.0,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE login_devices (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_agent VARCHAR(500) DEFAULT '',
    ip         VARCHAR(45) DEFAULT '',
    location   VARCHAR(255) DEFAULT '',
    last_login TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_login_devices_user_id ON login_devices (user_id);
