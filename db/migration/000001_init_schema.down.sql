ALTER TABLE "user_svc"."Users" DROP CONSTRAINT IF EXISTS "username" CASCADE;
ALTER TABLE "user_svc"."Users" DROP CONSTRAINT IF EXISTS "username" CASCADE;
ALTER TABLE "user_svc"."AccountTypes_Accounts" DROP CONSTRAINT IF EXISTS "AccountTypes_id" CASCADE;
ALTER TABLE "user_svc"."AccountTypes_Accounts" DROP CONSTRAINT IF EXISTS "Accounts_id" CASCADE;

DROP TABLE "user_svc"."Sessions" CASCADE;
DROP TABLE "user_svc"."Users" CASCADE;
DROP TABLE "user_svc"."AccountTypes_Accounts" CASCADE;
DROP TABLE "user_svc"."AccountTypes" CASCADE;
DROP TABLE "user_svc"."Accounts" CASCADE;

DROP SCHEMA "user_svc" CASCADE;
