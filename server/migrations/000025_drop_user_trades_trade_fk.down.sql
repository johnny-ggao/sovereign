-- Restore the FK. NOTE: this will fail if any user_trades.trade_id rows
-- point to trading_trades.id (because those will not exist in trades.id).
-- Before rolling back, either delete those rows or migrate them out.
ALTER TABLE user_trades
  ADD CONSTRAINT user_trades_trade_id_fkey
  FOREIGN KEY (trade_id) REFERENCES trades(id) ON DELETE SET NULL;
