-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, expires_at, user_id, updated_at, created_at)
VALUES (
	$1,
	$2,
	$3,
	now(),
	now()
)
RETURNING *;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = now(), updated_at = now() WHERE token = $1;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

