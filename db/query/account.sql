-- name: CreateAccount :one
INSERT INTO "user_svc"."Accounts" (
  username,
  email,
  password_hash,
  country_code,
  avatar_url
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id, username, email, country_code, created_at, updated_at;

-- name: GetAccountByID :one
SELECT
id,
username,
email, country_code,
avatar_url,
likes_count,
follows_count,
created_at,
updated_at
FROM "user_svc"."Accounts"
WHERE id = $1 LIMIT 1;

-- name: GetAccountByUsername :one
SELECT
id,
username,
email, country_code,
avatar_url,
likes_count,
follows_count,
created_at,
updated_at
FROM "user_svc"."Accounts"
WHERE username = $1 LIMIT 1;

-- name: GetAccountByAllParams :one
SELECT * FROM "user_svc"."Accounts"
WHERE username = $1 AND email = $2 AND country_code = $3 AND avatar_url = $4;

-- name: ListAccounts :many
SELECT id, username, country_code, created_at, updated_at FROM "user_svc"."Accounts"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAccount :one
UPDATE "user_svc"."Accounts"
SET
  username = COALESCE($2, username),
  email = COALESCE($3, email),
  country_code = COALESCE($4, country_code),
  avatar_url = COALESCE($5, avatar_url),
  likes_count = COALESCE($6, likes_count),
  follows_count = COALESCE($7, follows_count),
  updated_at = NOW()
WHERE id = $1
RETURNING id, username, country_code, created_at, updated_at;

-- name: UpdateAccountPassword :one
UPDATE "user_svc"."Accounts"
SET
  password_hash = COALESCE($2, password_hash)
WHERE id = $1
RETURNING id, username, created_at, updated_at;

-- name: DeleteAccount :exec
DELETE FROM "user_svc"."Accounts"
WHERE id = $1;
