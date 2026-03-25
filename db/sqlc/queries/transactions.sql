-- name: CreateTransaction :one
INSERT INTO transactions (
    account_id, category_id, amount, currency, type, description, notes, transacted_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetTransactionByIDForUser :one
SELECT t.*
FROM transactions t
JOIN accounts a ON a.id = t.account_id
WHERE t.id = $1
  AND a.user_id = $2
  AND a.deleted_at IS NULL
  AND t.deleted_at IS NULL;

-- name: GetTransactionByIDForUserForUpdate :one
SELECT t.*
FROM transactions t
JOIN accounts a ON a.id = t.account_id
WHERE t.id = $1
  AND a.user_id = $2
  AND a.deleted_at IS NULL
  AND t.deleted_at IS NULL
FOR UPDATE;

-- name: ListTransactionsForUser :many
SELECT t.*
FROM transactions t
JOIN accounts a ON a.id = t.account_id
WHERE a.user_id = sqlc.arg(user_id)
  AND a.deleted_at IS NULL
  AND t.deleted_at IS NULL
  AND (sqlc.narg(account_id)::bigint IS NULL OR t.account_id = sqlc.narg(account_id)::bigint)
  AND (sqlc.narg(category_id)::bigint IS NULL OR t.category_id = sqlc.narg(category_id)::bigint)
  AND (sqlc.narg(type)::text IS NULL OR t.type = sqlc.narg(type)::text)
  AND (sqlc.narg(from_date)::date IS NULL OR t.transacted_at >= sqlc.narg(from_date)::date)
  AND (sqlc.narg(to_date)::date IS NULL OR t.transacted_at <= sqlc.narg(to_date)::date)
ORDER BY t.transacted_at DESC, t.id DESC
LIMIT sqlc.arg(limit_rows)
OFFSET sqlc.arg(offset_rows);

-- name: UpdateTransactionByIDForUser :one
UPDATE transactions t
SET amount = COALESCE($3, t.amount),
    category_id = COALESCE($4, t.category_id),
    notes = COALESCE($5, t.notes),
    updated_at = NOW()
FROM accounts a
WHERE t.id = $1
  AND a.id = t.account_id
  AND a.user_id = $2
  AND a.deleted_at IS NULL
  AND t.deleted_at IS NULL
RETURNING t.*;

-- name: SoftDeleteTransactionByIDForUser :execrows
UPDATE transactions t
SET deleted_at = NOW(),
    updated_at = NOW()
FROM accounts a
WHERE t.id = $1
  AND a.id = t.account_id
  AND a.user_id = $2
  AND a.deleted_at IS NULL
  AND t.deleted_at IS NULL;

-- name: UpdateAccountBalanceDeltaByIDForUser :one
UPDATE accounts
SET balance = balance + $3,
    updated_at = NOW()
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
RETURNING *;

-- name: CategoryAccessibleForUser :one
SELECT id
FROM categories
WHERE id = $1
  AND (user_id = $2 OR is_system = TRUE)
LIMIT 1;
