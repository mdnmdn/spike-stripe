-- 0004_add_refund_date.sql
-- Add refund_date field to transactions table
ALTER TABLE transactions ADD COLUMN refund_date TEXT;

-- Create index for refund date to support queries on refunded transactions
CREATE INDEX IF NOT EXISTS idx_transactions_refund_date ON transactions(refund_date);