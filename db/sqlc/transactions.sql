-- Transactions

-- name: CreateTransaction :one
INSERT INTO transactions (account_id, amount, description, transaction_type)
VALUES ($1, $2, $3, $4)
RETURNING id, account_id, amount, description, transaction_type, created_at;

-- name: ListTransactionsByAccountID :many
SELECT id, account_id, amount, description, transaction_type, created_at
FROM transactions
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

