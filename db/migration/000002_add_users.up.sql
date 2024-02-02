CREATE TABLE "user_svc"."Users" (
  "id" BIGSERIAL PRIMARY KEY,
  "username" VARCHAR(255) UNIQUE NOT NULL,
  "full_name" VARCHAR(255) NOT NULL,
  "email" VARCHAR(255) UNIQUE NOT NULL,
  "password_hash" VARCHAR(255) NOT NULL,
  "password_salt" VARCHAR(255) NOT NULL,
  "country_code" VARCHAR(10) NOT NULL,
  "role_id" BIGINT,
  "status" VARCHAR(50),
  "last_login_at" TIMESTAMPTZ DEFAULT '0001-01-01 00:00:00Z',
  "username_changed_at" TIMESTAMPTZ NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "email_changed_at" TIMESTAMPTZ NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "password_changed_at" TIMESTAMPTZ NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now()),
  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE INDEX "idx_user_id" ON "user_svc"."Users" ("id");

CREATE INDEX "idx_user_username" ON "user_svc"."Users" ("username");

CREATE INDEX "idx_users_email" ON "user_svc"."Users" ("email");

ALTER TABLE "user_svc"."Accounts" ADD FOREIGN KEY ("owner") REFERENCES "user_svc"."Users" ("username");