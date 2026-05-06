ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS review_status     varchar(32),
  ADD COLUMN IF NOT EXISTS reviewed_by       uuid,
  ADD COLUMN IF NOT EXISTS reviewed_at       timestamptz,
  ADD COLUMN IF NOT EXISTS reject_reason     varchar(500),
  ADD COLUMN IF NOT EXISTS submit_attempts   int NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_submit_error text,
  ADD COLUMN IF NOT EXISTS last_submit_at    timestamptz;

CREATE INDEX IF NOT EXISTS idx_tx_review_status ON transactions(review_status)
  WHERE type = 'withdraw';

-- Backfill: historical withdraw rows are treated as already submitted directly.
UPDATE transactions
   SET review_status = 'submitted'
 WHERE type = 'withdraw' AND review_status IS NULL;
