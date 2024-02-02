-- name: CreateUser :one
INSERT INTO "user_svc"."Users" (
 username,
 full_name,
 email,
 password_hash,
 password_salt,
 country_code,
 role_id,
 status
) VALUES (
 $1, $2, $3, $4, $5 , $6, $7, $8
)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM "user_svc"."Users"
WHERE username = $1 LIMIT 1;

-- name: GetUserByID :one
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

-- name: UpdateUserEmail :one
UPDATE "user_svc"."Users"
SET email = COALESCE($2, email),
    email_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING email, email_changed_at, updated_at;

-- name: UpdateUserPassword :one
UPDATE "user_svc"."Users"
SET password_hash = COALESCE($2, password_hash),
    password_salt = COALESCE($3, password_salt),
    password_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING password_hash, password_salt, password_changed_at, updated_at;

-- name: UpdateUsername :one
UPDATE "user_svc"."Users"
SET username = COALESCE($2, username),
    username_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING username, username_changed_at, updated_at;

-- name: UpdateUser :one
UPDATE "user_svc"."Users"
SET username = COALESCE($2, username),
    full_name = COALESCE($3, full_name),
    country_code = COALESCE($4, country_code),
    role_id = COALESCE($5, role_id),
    status = COALESCE($6, status),
    updated_at = NOW()
WHERE id = $1
RETURNING username, full_name, country_code, role_id, status, last_login_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM "user_svc"."Users"
WHERE id = $1;
