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