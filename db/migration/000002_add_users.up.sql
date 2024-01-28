CREATE TABLE "user_svc"."Users" (
  "id" bigserial PRIMARY KEY,
  "username" varchar UNIQUE NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_hash" varchar NOT NULL,
  "country_code" varchar NOT NULL,
  "username_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "email_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX "idx_user_id" ON "user_svc"."Users" ("id");

CREATE INDEX "idx_username" ON "user_svc"."Users" ("username");

ALTER TABLE "user_svc"."Accounts" ADD FOREIGN KEY ("owner") REFERENCES "user_svc"."Users" ("username");