-- name: AddAccountTypeToAccount :exec
INSERT INTO "user_svc"."AccountTypes_Accounts" ("Accounts_id", "AccountTypes_id")
VALUES ($1, $2);

-- name: GetAccountTypesForAccount :many
SELECT at.* FROM "user_svc"."AccountTypes" at
JOIN "user_svc"."AccountTypes_Accounts" aat ON at.id = aat."AccountTypes_id"
WHERE aat."Accounts_id" = $1;

-- name: GetAccountsForAccountType :many
SELECT ac.* FROM "user_svc"."Accounts" ac
JOIN "user_svc"."AccountTypes_Accounts" aat ON ac.id = aat."Accounts_id"
WHERE aat."AccountTypes_id" = $1;

-- name: RemoveAccountTypeFromAccount :exec
DELETE FROM "user_svc"."AccountTypes_Accounts"
WHERE "Accounts_id" = $1 AND "AccountTypes_id" = $2;

-- name: RemoveAllRelationshipsForAccountAccountType :exec
DELETE FROM "user_svc"."AccountTypes_Accounts"
WHERE "Accounts_id" = $1;

-- name: RemoveAllRelationshipsForAccountTypeAccount :exec
DELETE FROM "user_svc"."AccountTypes_Accounts"
WHERE "AccountTypes_id" = $1;
