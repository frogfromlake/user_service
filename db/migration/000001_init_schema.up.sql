CREATE SCHEMA "user_svc";

CREATE TABLE "user_svc"."Users" (
  "id" bigserial PRIMARY KEY,
  "username" varchar UNIQUE NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_hash" varchar NOT NULL,
  "password_salt" varchar NOT NULL,
  "country_code" varchar NOT NULL,
  "role_id" bigint,
  "status" varchar,
  "last_login_at" timestamptz NOT NULL DEFAULT('0001-01-01 00:00:00Z'),
  "username_changed_at" timestamptz NOT NULL DEFAULT('0001-01-01 00:00:00Z'),
  "email_changed_at" timestamptz NOT NULL DEFAULT('0001-01-01 00:00:00Z'),
  "password_changed_at" timestamptz NOT NULL DEFAULT('0001-01-01 00:00:00Z'),
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "user_svc"."Sessions" (
  "id" uuid PRIMARY KEY,
  "username" varchar NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX "idx_user_id" ON "user_svc"."Users" ("id");

CREATE INDEX "idx_user_username" ON "user_svc"."Users" ("username");

CREATE INDEX "idx_users_email" ON "user_svc"."Users" ("email");

CREATE INDEX "idx_session_username" ON "user_svc"."Sessions" ("username");

ALTER TABLE "user_svc"."Sessions" ADD FOREIGN KEY ("username") REFERENCES "user_svc"."Users" ("username");