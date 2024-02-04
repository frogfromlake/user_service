CREATE SCHEMA "user_svc";

CREATE TABLE "user_svc"."Accounts" (
  "id" bigserial PRIMARY KEY,
  "account_type" int UNIQUE NOT NULL,
  "owner" varchar NOT NULL,
  "avatar_uri" varchar,
  "plays" bigint NOT NULL DEFAULT 0,
  "likes" bigint NOT NULL DEFAULT 0,
  "follows" bigint NOT NULL DEFAULT 0,
  "shares" bigint NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "user_svc"."AccountTypes" (
  "id" serial PRIMARY KEY,
  "type" varchar UNIQUE NOT NULL,
  "permissions" jsonb NOT NULL,
  "is_artist" boolean NOT NULL DEFAULT false,
  "is_producer" boolean NOT NULL DEFAULT false,
  "is_writer" boolean NOT NULL DEFAULT false,
  "is_label" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "user_svc"."AccountTypes_Accounts" (
  "AccountTypes_id" serial,
  "Accounts_id" bigserial,
  PRIMARY KEY ("AccountTypes_id", "Accounts_id")
);

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
  "last_login_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "username_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "email_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
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

CREATE INDEX "idx_acc_id" ON "user_svc"."Accounts" ("id");

CREATE INDEX "idx_acc_owner" ON "user_svc"."Accounts" ("owner");

CREATE INDEX "idx_accType_id" ON "user_svc"."AccountTypes" ("id");

CREATE INDEX "idx_user_id" ON "user_svc"."Users" ("id");

CREATE INDEX "idx_user_username" ON "user_svc"."Users" ("username");

CREATE INDEX "idx_users_email" ON "user_svc"."Users" ("email");

CREATE INDEX "idx_session_username" ON "user_svc"."Sessions" ("username");

ALTER TABLE "user_svc"."AccountTypes_Accounts" ADD FOREIGN KEY ("AccountTypes_id") REFERENCES "user_svc"."AccountTypes" ("id");

ALTER TABLE "user_svc"."AccountTypes_Accounts" ADD FOREIGN KEY ("Accounts_id") REFERENCES "user_svc"."Accounts" ("id");

ALTER TABLE "user_svc"."Accounts" ADD CONSTRAINT "unique_account" UNIQUE ("owner", "account_type");

ALTER TABLE "user_svc"."Sessions" ADD FOREIGN KEY ("username") REFERENCES "user_svc"."Users" ("username");