-- name: CreateAccount :one
INSERT INTO "user_svc"."Accounts" (
 owner,
 account_type,
 avatar_uri
) VALUES (
 $1, $2, $3
)
RETURNING id, owner, account_type, avatar_uri, created_at, updated_at;

-- name: GetAccountByID :one
SELECT * FROM "user_svc"."Accounts"
WHERE id = $1 LIMIT 1;

-- name: GetAccountByOwner :one
SELECT * FROM "user_svc"."Accounts"
WHERE owner = $1 LIMIT 1;

-- name: ListAccounts :many
SELECT id, owner, account_type, created_at, updated_at FROM "user_svc"."Accounts"
WHERE owner = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: UpdateAccount :one
UPDATE "user_svc"."Accounts"
SET
 owner = COALESCE($2, owner),
 account_type = COALESCE($3, account_type),
 avatar_uri = COALESCE($4, avatar_uri),
 plays = COALESCE($5, plays),
 likes = COALESCE($6, likes),
 follows = COALESCE($7, follows),
 shares = COALESCE($8, shares),
 updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM "user_svc"."Accounts"
WHERE id = $1;
