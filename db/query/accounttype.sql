-- name: CreateAccountType :one
INSERT INTO "user_svc"."AccountTypes" (
 type,
 permissions,
 is_artist,
 is_producer,
 is_writer,
 is_label
) VALUES (
 $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetAccountType :one
SELECT * FROM "user_svc"."AccountTypes"
WHERE id = $1 LIMIT 1;

-- name: ListAccountTypes :many
SELECT * FROM "user_svc"."AccountTypes"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAccountType :one
UPDATE "user_svc"."AccountTypes"
SET 
 type = COALESCE($2, type),
 permissions = COALESCE($3, permissions),
 is_artist = COALESCE($4, is_artist),
 is_producer = COALESCE($5, is_producer),
 is_writer = COALESCE($6, is_writer),
 is_label = COALESCE($7, is_label),
 updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAccountType :exec
DELETE FROM "user_svc"."AccountTypes"
WHERE id = $1;
