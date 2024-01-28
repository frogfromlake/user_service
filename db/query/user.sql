-- name: CreateUser :one
INSERT INTO "user_svc"."Users" (
 username,
 full_name,
 email,
 password_hash,
 country_code
) VALUES (
 $1, $2, $3, $4, $5
)
RETURNING id, username, full_name, email, country_code, created_at;

-- name: GetUserByUsername :one
SELECT
 id,
 username,
 full_name,
 email,
 country_code,
 username_changed_at,
 email_changed_at,
 password_changed_at,
 created_at,
 updated_at
FROM "user_svc"."Users"
WHERE username = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT
 id,
 username,
 full_name,
 email,
 country_code,
 username_changed_at,
 email_changed_at,
 password_changed_at,
 created_at,
 updated_at
FROM "user_svc"."Users"
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT
 id,
 username,
 full_name,
 email,
 country_code,
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
    password_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING password_hash, password_changed_at, updated_at;

-- name: UpdateUsername :one
UPDATE "user_svc"."Users"
SET username = COALESCE($2, username),
    username_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING username, username_changed_at, updated_at;

-- name: UpdateUser :one
UPDATE "user_svc"."Users"
SET full_name = COALESCE($2, full_name),
    country_code = COALESCE($3, country_code),
    updated_at = NOW()
WHERE id = $1
RETURNING full_name, country_code, updated_at;

-- name: DeleteUser :exec
DELETE FROM "user_svc"."Users"
WHERE id = $1;
