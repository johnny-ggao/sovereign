ALTER TABLE transactions
  ADD COLUMN review_status     varchar(32),
  ADD COLUMN reviewed_by       uuid,
  ADD COLUMN reviewed_at       timestamptz,
  ADD COLUMN reject_reason     varchar(500),
  ADD COLUMN submit_attempts   int NOT NULL DEFAULT 0,
  ADD COLUMN last_submit_error text,
  ADD COLUMN last_submit_at    timestamptz;

CREATE INDEX idx_tx_review_status ON transactions(review_status)
  WHERE type = 'withdraw';

-- Backfill: historical withdraw rows are treated as already submitted directly.
UPDATE transactions
   SET review_status = 'submitted'
 WHERE type = 'withdraw' AND review_status IS NULL;
