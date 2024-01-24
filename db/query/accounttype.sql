-- name: CreateAccountType :one
INSERT INTO "user_service"."AccountTypes" (
  description,
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
SELECT * FROM "user_service"."AccountTypes"
WHERE id = $1 LIMIT 1;

-- name: GetAccountTypeByAllParams :one
SELECT * FROM "user_service"."AccountTypes"
WHERE description = $1 AND permissions = $2 AND is_artist = $3 AND is_producer = $4 AND is_writer = $5 AND is_label = $6 LIMIT 1;

-- name: ListAccountTypes :many
SELECT * FROM "user_service"."AccountTypes"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAccountType :one
UPDATE "user_service"."AccountTypes"
SET 
  description = COALESCE($2, description),
  permissions = COALESCE($3, permissions),
  is_artist = COALESCE($4, is_artist),
  is_producer = COALESCE($5, is_producer),
  is_writer = COALESCE($6, is_writer),
  is_label = COALESCE($7, is_label),
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAccountType :exec
DELETE FROM "user_service"."AccountTypes"
WHERE id = $1;
