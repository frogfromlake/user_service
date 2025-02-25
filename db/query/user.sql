-- name: CreateUser :one
INSERT INTO "user_svc"."Users" (
 username,
 full_name,
 email,
 password_hash,
 password_salt,
 country_code,
 role_id,
 status,
 last_login_at,
 username_changed_at,
 email_changed_at,
 password_changed_at,
 created_at,
 updated_at
) VALUES (
 $1, $2, $3, $4, $5 , $6, $7, $8,
    '0001-01-01 00:00:00Z',
    '0001-01-01 00:00:00Z',
    '0001-01-01 00:00:00Z',
    '0001-01-01 00:00:00Z',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
RETURNING *;

-- name: GetUserByValue :one
SELECT * FROM "user_svc"."Users"
WHERE username = $1 LIMIT 1;

-- name: GetUserById :one
SELECT * FROM "user_svc"."Users"
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT
 id,
 username,
 full_name,
 email,
 country_code,
 role_id,
 status,
 last_login_at,
 created_at,
 updated_at
FROM "user_svc"."Users"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE "user_svc"."Users"
SET 
    username = COALESCE(sqlc.narg(username), username),
    full_name = COALESCE(sqlc.narg(full_name), full_name),
    email = COALESCE(sqlc.narg(email), email),
    password_hash = COALESCE(sqlc.narg(password_hash), password_hash),
    password_salt = COALESCE(sqlc.narg(password_salt), password_salt),
    country_code = COALESCE(sqlc.narg(country_code), country_code),
    role_id = COALESCE(sqlc.narg(role_id), role_id),
    status = COALESCE(sqlc.narg(status), status),
    last_login_at = COALESCE(sqlc.narg(last_login_at), last_login_at),
    username_changed_at = COALESCE(sqlc.narg(username_changed_at), username_changed_at),
    email_changed_at = COALESCE(sqlc.narg(email_changed_at), email_changed_at),
    password_changed_at = COALESCE(sqlc.narg(password_changed_at), password_changed_at),
    created_at = COALESCE(sqlc.narg(created_at), created_at),
    updated_at = NOW()
WHERE "user_svc"."Users".id = sqlc.arg(id)
RETURNING *;

-- name: DeleteUserById :exec
DELETE FROM "user_svc"."Users"
WHERE id = $1;

-- name: DeleteUserByValue :exec
DELETE FROM "user_svc"."Users"
WHERE username = $1;
