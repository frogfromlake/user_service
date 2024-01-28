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

CREATE INDEX "idx_acc_id" ON "user_svc"."Accounts" ("id");

CREATE INDEX "idx_acc_owner" ON "user_svc"."Accounts" ("owner");

CREATE INDEX "idx_accType_id" ON "user_svc"."AccountTypes" ("id");

CREATE TABLE "user_svc"."AccountTypes_Accounts" (
  "AccountTypes_id" serial,
  "Accounts_id" bigserial,
  PRIMARY KEY ("AccountTypes_id", "Accounts_id")
);

ALTER TABLE "user_svc"."AccountTypes_Accounts" ADD FOREIGN KEY ("AccountTypes_id") REFERENCES "user_svc"."AccountTypes" ("id");

ALTER TABLE "user_svc"."AccountTypes_Accounts" ADD FOREIGN KEY ("Accounts_id") REFERENCES "user_svc"."Accounts" ("id");

ALTER TABLE "user_svc"."Accounts" ADD CONSTRAINT "unique_account" UNIQUE ("owner", "account_type");
