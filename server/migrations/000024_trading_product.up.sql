-- 1. New: per-user product type marker
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS investment_type varchar(20) NOT NULL DEFAULT 'arbitrage';

-- 2. New: product_type on user-facing tables
ALTER TABLE investments
  ADD COLUMN IF NOT EXISTS product_type varchar(20) NOT NULL DEFAULT 'arbitrage';
CREATE INDEX IF NOT EXISTS idx_investments_product ON investments(product_type, status);

ALTER TABLE user_trades
  ADD COLUMN IF NOT EXISTS product_type varchar(20) NOT NULL DEFAULT 'arbitrage';
CREATE INDEX IF NOT EXISTS idx_user_trades_product ON user_trades(product_type);

ALTER TABLE settlements
  ADD COLUMN IF NOT EXISTS product_type varchar(20) NOT NULL DEFAULT 'arbitrage';
CREATE INDEX IF NOT EXISTS idx_settlements_product ON settlements(product_type, period);

-- 3. Wallet earnings split (backfill BEFORE dropping the old column)
ALTER TABLE wallets
  ADD COLUMN IF NOT EXISTS earnings_arbitrage decimal(28,18) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS earnings_trading   decimal(28,18) NOT NULL DEFAULT 0;
UPDATE wallets SET earnings_arbitrage = COALESCE(earnings, 0) WHERE earnings_arbitrage = 0;
ALTER TABLE wallets DROP COLUMN IF EXISTS earnings;

-- 4. Audit table for tagging changes
CREATE TABLE IF NOT EXISTS user_product_change_logs (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid NOT NULL,
  from_type   varchar(20) NOT NULL,
  to_type     varchar(20) NOT NULL,
  admin_id    uuid NOT NULL,
  admin_email varchar(255) NOT NULL,
  reason      varchar(500) NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_upcl_user ON user_product_change_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_upcl_created ON user_product_change_logs(created_at);

-- 5. Trading fund trades (mirrors trades schema)
CREATE TABLE IF NOT EXISTS trading_trades (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  pair          varchar(20) NOT NULL,
  buy_exchange  varchar(20),
  sell_exchange varchar(20),
  buy_price     decimal(28,8),
  sell_price    decimal(28,8),
  amount        decimal(28,18) NOT NULL,
  premium_pct   decimal(8,4) DEFAULT 0,
  pnl           decimal(28,18) NOT NULL,
  fee           decimal(28,18) DEFAULT 0,
  source        varchar(20) NOT NULL DEFAULT 'import',
  executed_at   timestamptz NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_trading_trades_executed_at ON trading_trades(executed_at);
