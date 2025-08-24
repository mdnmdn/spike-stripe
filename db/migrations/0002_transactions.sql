-- 0002_transactions.sql
-- Transactions table for Stripe payment tracking
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    product_name TEXT NOT NULL,
    amount INTEGER NOT NULL, -- price in cents
    stripe_session_id TEXT,
    stripe_payment_intent_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, completed, failed, cancelled
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_transactions_stripe_session_id ON transactions(stripe_session_id);