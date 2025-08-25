-- name: CreateTransaction :exec
INSERT INTO transactions (id, user_id, product_id, product_name, amount, stripe_session_id, stripe_payment_intent_id, status, created_at, updated_at, refund_date)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTransaction :one
SELECT id, user_id, product_id, product_name, amount, stripe_session_id, stripe_payment_intent_id, status, created_at, updated_at, refund_date
FROM transactions
WHERE id = ?
LIMIT 1;

-- name: GetTransactionByStripeSessionID :one
SELECT id, user_id, product_id, product_name, amount, stripe_session_id, stripe_payment_intent_id, status, created_at, updated_at, refund_date
FROM transactions
WHERE stripe_session_id = ?
LIMIT 1;

-- name: ListTransactionsByUserID :many
SELECT id, user_id, product_id, product_name, amount, stripe_session_id, stripe_payment_intent_id, status, created_at, updated_at, refund_date
FROM transactions
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAllTransactions :many
SELECT id, user_id, product_id, product_name, amount, stripe_session_id, stripe_payment_intent_id, status, created_at, updated_at, refund_date
FROM transactions
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateTransactionStatus :exec
UPDATE transactions 
SET status = ?, updated_at = ?
WHERE id = ?;

-- name: UpdateTransactionWithStripeData :exec
UPDATE transactions 
SET stripe_payment_intent_id = ?, status = ?, updated_at = ?
WHERE stripe_session_id = ?;

-- name: UpdateTransactionByPaymentIntentID :exec
UPDATE transactions 
SET status = ?, updated_at = ?
WHERE stripe_payment_intent_id = ?;

-- name: UpdateTransactionByPaymentIntentIDWithRefundDate :exec
UPDATE transactions 
SET status = ?, updated_at = ?, refund_date = ?
WHERE stripe_payment_intent_id = ?;