CREATE SCHEMA "user_service";

CREATE TABLE "user_service"."Accounts" (
  "id" bigserial PRIMARY KEY,
  "username" varchar UNIQUE NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_hash" varchar NOT NULL,
  "country_code" varchar NOT NULL,
  "avatar_url" varchar,
  "likes_count" bigint NOT NULL DEFAULT 0,
  "follows_count" bigint NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "user_service"."AccountTypes" (
  "id" bigserial PRIMARY KEY,
  "description" text,
  "permissions" jsonb,
  "is_artist" boolean NOT NULL DEFAULT false,
  "is_producer" boolean NOT NULL DEFAULT false,
  "is_writer" boolean NOT NULL DEFAULT false,
  "is_label" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX "idx_acc_id" ON "user_service"."Accounts" ("id");

CREATE INDEX "idx_acc_username" ON "user_service"."Accounts" ("username");

CREATE INDEX "idx_acc_email" ON "user_service"."Accounts" ("email");

CREATE INDEX "idx_accType_id" ON "user_service"."AccountTypes" ("id");

CREATE TABLE "user_service"."Accounts_AccountTypes" (
  "Accounts_id" bigserial,
  "AccountTypes_id" bigserial,
  PRIMARY KEY ("Accounts_id", "AccountTypes_id")
);

ALTER TABLE "user_service"."Accounts_AccountTypes" ADD FOREIGN KEY ("Accounts_id") REFERENCES "user_service"."Accounts" ("id");

ALTER TABLE "user_service"."Accounts_AccountTypes" ADD FOREIGN KEY ("AccountTypes_id") REFERENCES "user_service"."AccountTypes" ("id");