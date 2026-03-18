-- Accounts

-- name: CreateAccount :one
INSERT INTO accounts (user_id, account_type, balance, currency)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, account_type, balance, currency, created_at, updated_at;

-- name: GetAccountsByUserID :many
SELECT id, user_id, account_type, balance, currency, created_at, updated_at
FROM accounts
WHERE user_id = $1
ORDER BY created_at DESC;

