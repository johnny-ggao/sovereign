DROP TABLE IF EXISTS trading_trades;
DROP TABLE IF EXISTS user_product_change_logs;

ALTER TABLE wallets ADD COLUMN IF NOT EXISTS earnings decimal(28,18) NOT NULL DEFAULT 0;
UPDATE wallets SET earnings = earnings_arbitrage + earnings_trading;
ALTER TABLE wallets DROP COLUMN IF EXISTS earnings_arbitrage;
ALTER TABLE wallets DROP COLUMN IF EXISTS earnings_trading;

DROP INDEX IF EXISTS idx_settlements_product;
ALTER TABLE settlements DROP COLUMN IF EXISTS product_type;

DROP INDEX IF EXISTS idx_user_trades_product;
ALTER TABLE user_trades DROP COLUMN IF EXISTS product_type;

DROP INDEX IF EXISTS idx_investments_product;
ALTER TABLE investments DROP COLUMN IF EXISTS product_type;

ALTER TABLE users DROP COLUMN IF EXISTS investment_type;
