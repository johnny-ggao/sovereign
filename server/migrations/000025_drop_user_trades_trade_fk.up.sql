-- user_trades.trade_id can now point to either trades.id (arbitrage) or
-- trading_trades.id (trading), depending on product_type. Postgres can't
-- model a polymorphic FK, and product_type already disambiguates the
-- source table, so drop the original FK constraint.
ALTER TABLE user_trades DROP CONSTRAINT IF EXISTS user_trades_trade_id_fkey;
