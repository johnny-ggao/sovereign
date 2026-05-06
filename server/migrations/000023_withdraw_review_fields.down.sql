DROP INDEX IF EXISTS idx_tx_review_status;
ALTER TABLE transactions
  DROP COLUMN IF EXISTS review_status,
  DROP COLUMN IF EXISTS reviewed_by,
  DROP COLUMN IF EXISTS reviewed_at,
  DROP COLUMN IF EXISTS reject_reason,
  DROP COLUMN IF EXISTS submit_attempts,
  DROP COLUMN IF EXISTS last_submit_error,
  DROP COLUMN IF EXISTS last_submit_at;
