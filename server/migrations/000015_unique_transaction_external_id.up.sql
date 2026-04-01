-- 先清理重复的 external_id 记录（保留最早的）
DELETE FROM transactions a USING transactions b
WHERE a.id > b.id AND a.external_id = b.external_id AND a.external_id != '';

-- 添加唯一索引（排除空字符串）
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_external_id_unique
ON transactions (external_id) WHERE external_id != '';
