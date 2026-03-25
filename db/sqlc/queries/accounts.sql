-- name: CreateAccount :one
INSERT INTO accounts (user_id, name, account_type, balance, currency)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListAccountsByUser :many
SELECT *
FROM accounts
WHERE user_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetAccountByIDForUser :one
SELECT *
FROM accounts
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: GetAccountByIDForUserForUpdate :one
SELECT *
FROM accounts
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
FOR UPDATE;

-- name: UpdateAccountByIDForUser :one
UPDATE accounts
SET name = COALESCE(sqlc.narg(name)::text, name),
    account_type = COALESCE(sqlc.narg(account_type)::text, account_type),
    currency = COALESCE(sqlc.narg(currency)::text, currency),
    updated_at = NOW()
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteAccountByIDForUser :execrows
UPDATE accounts
SET deleted_at = NOW(),
    is_active = FALSE,
    updated_at = NOW()
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL;
