-- name: CreateAccount :one
INSERT INTO "user_svc"."Accounts" (
 owner,
 avatar_url
) VALUES (
 $1, $2
)
RETURNING id, owner, avatar_url, created_at, updated_at;

-- name: GetAccountByID :one
SELECT
id,
owner,
avatar_url,
plays,
likes,
follows,
shares,
created_at,
updated_at
FROM "user_svc"."Accounts"
WHERE id = $1 LIMIT 1;

-- name: GetAccountByOwner :one
SELECT
id,
owner,
avatar_url,
plays,
likes,
follows,
shares,
created_at,
updated_at
FROM "user_svc"."Accounts"
WHERE owner = $1 LIMIT 1;

-- name: ListAccounts :many
SELECT id, owner, created_at, updated_at FROM "user_svc"."Accounts"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAccount :one
UPDATE "user_svc"."Accounts"
SET
 owner = COALESCE($2, owner),
 avatar_url = COALESCE($3, avatar_url),
 plays = COALESCE($4, plays),
 likes = COALESCE($5, likes),
 follows = COALESCE($6, follows),
 shares = COALESCE($7, shares),
 updated_at = NOW()
WHERE id = $1
RETURNING *;

-- -- name: UpdateAccountPassword :one
-- UPDATE "user_svc"."Accounts"
-- SET
--   password_hash = COALESCE($2, password_hash)
-- WHERE id = $1
-- RETURNING id, username, created_at, updated_at;

-- name: DeleteAccount :exec
DELETE FROM "user_svc"."Accounts"
WHERE id = $1;
