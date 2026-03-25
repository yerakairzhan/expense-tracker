-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, currency)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
  AND deleted_at IS NULL
  AND is_active = TRUE;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
  AND deleted_at IS NULL
  AND is_active = TRUE;

-- name: UpdateUserProfile :one
UPDATE users
SET name = COALESCE(sqlc.narg(name)::text, name),
    currency = COALESCE(sqlc.narg(currency)::text, currency),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET password_hash = $2,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: InsertRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListValidRefreshTokens :many
SELECT *
FROM refresh_tokens
WHERE revoked = FALSE
  AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: ListValidRefreshTokensByUser :many
SELECT *
FROM refresh_tokens
WHERE user_id = $1
  AND revoked = FALSE
  AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: RevokeRefreshTokenByID :execrows
UPDATE refresh_tokens
SET revoked = TRUE
WHERE id = $1
  AND revoked = FALSE;

-- name: RevokeRefreshTokenByIDForUser :execrows
UPDATE refresh_tokens
SET revoked = TRUE
WHERE id = $1
  AND user_id = $2
  AND revoked = FALSE;
