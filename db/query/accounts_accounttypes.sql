-- name: AddAccountTypeToAccount :exec
INSERT INTO "user_svc"."Accounts_AccountTypes" ("Accounts_id", "AccountTypes_id")
VALUES ($1, $2);

-- name: GetAccountTypesForAccount :many
SELECT at.* FROM "user_svc"."AccountTypes" at
JOIN "user_svc"."Accounts_AccountTypes" aat ON at.id = aat."AccountTypes_id"
WHERE aat."Accounts_id" = $1;

-- name: GetAccountTypeIDsForAccount :many
SELECT at.id, at.created_at, at.updated_at FROM "user_svc"."AccountTypes" at
JOIN "user_svc"."Accounts_AccountTypes" aat ON at.id = aat."AccountTypes_id"
WHERE aat."Accounts_id" = $1;

-- name: GetAccountsForAccountType :many
SELECT ac.id, ac.username, ac.email, ac.country_code, ac.avatar_url, ac.likes_count, ac.follows_count, ac.created_at, ac.updated_at FROM "user_svc"."Accounts" ac
JOIN "user_svc"."Accounts_AccountTypes" aat ON ac.id = aat."Accounts_id"
WHERE aat."AccountTypes_id" = $1;

-- name: RemoveAccountTypeFromAccount :exec
DELETE FROM "user_svc"."Accounts_AccountTypes"
WHERE "Accounts_id" = $1 AND "AccountTypes_id" = $2;

-- name: RemoveAllRelationshipsForAccountAccountType :exec
DELETE FROM "user_svc"."Accounts_AccountTypes"
WHERE "Accounts_id" = $1;

-- name: RemoveAllRelationshipsForAccountTypeAccount :exec
DELETE FROM "user_svc"."Accounts_AccountTypes"
WHERE "AccountTypes_id" = $1;
