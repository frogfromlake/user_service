ALTER TABLE "user_svc"."Users" DROP CONSTRAINT IF EXISTS "username" CASCADE;
ALTER TABLE "user_svc"."Users" DROP CONSTRAINT IF EXISTS "username" CASCADE;

DROP TABLE "user_svc"."Sessions" CASCADE;
DROP TABLE "user_svc"."Users" CASCADE;

DROP SCHEMA "user_svc" CASCADE;
