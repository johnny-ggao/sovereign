CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          VARCHAR(255) UNIQUE NOT NULL,
    password_hash  VARCHAR(255) NOT NULL,
    full_name      VARCHAR(255) DEFAULT '',
    phone          VARCHAR(50) DEFAULT '',
    language       VARCHAR(5) DEFAULT 'ko',
    kyc_status     VARCHAR(20) DEFAULT 'pending',
    two_fa_secret  TEXT DEFAULT '',
    two_fa_enabled BOOLEAN DEFAULT false,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_kyc_status ON users (kyc_status);
