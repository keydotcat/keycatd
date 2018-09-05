DROP TABLE IF EXISTS "user" CASCADE;
CREATE TABLE "user" (
	"id" TEXT NOT NULL,
	"email" TEXT NOT NULL,
	"unconfirmed_email" TEXT NULL,
	"hash_pass" BYTEA NOT NULL,
	"full_name" TEXT NOT NULL,
	"confirmed_at" TIMESTAMP WITH TIME ZONE NULL,
	"locked_at" TIMESTAMP WITH TIME ZONE NULL,
	"sign_in_count" INT NULL DEFAULT 0,
	"failed_attempts" INT NULL DEFAULT 0,
	"public_key" BYTEA NOT NULL,
	"key" BYTEA NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE,
	"updated_at" TIMESTAMP WITH TIME ZONE,
	CONSTRAINT "pk_user" PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "idx_user_email" ON "user" ("email");

DROP TABLE IF EXISTS "team" CASCADE;
CREATE TABLE "team" (
	"id" TEXT NOT NULL,
	"name" TEXT NOT NULL,
	"owner" TEXT NOT NULL,
	"primary" BOOL NOT NULL,
	"size" INT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "pk_team" PRIMARY KEY ("id" ),
	CONSTRAINT "fk_team_user" FOREIGN KEY ("owner") REFERENCES "user" ON DELETE CASCADE
);

DROP TABLE IF EXISTS "invite" CASCADE;
CREATE TABLE "invite" (
	"team" TEXT NOT NULL,
	"email" TEXT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "pk_invite" PRIMARY KEY ("team", "email"),
	CONSTRAINT "fk_invite_eam" FOREIGN KEY ("team") REFERENCES "team" ON DELETE CASCADE
);
CREATE UNIQUE INDEX "idx_invite_email" ON "invite" ("email");

DROP TABLE IF EXISTS "team_user" CASCADE;
CREATE TABLE "team_user" (
	"team" TEXT NOT NULL,
	"user" TEXT NOT NULL,
	"admin" BOOL NOT NULL,
	"access_required" BOOL NOT NULL,
	CONSTRAINT "pk_team_user" PRIMARY KEY ("team", "user"),
	CONSTRAINT "fk_team_user_user" FOREIGN KEY ("user") REFERENCES "user" ON DELETE CASCADE,
	CONSTRAINT "fk_team_user_team" FOREIGN KEY ("team") REFERENCES "team" ON DELETE CASCADE
);
CREATE INDEX "idx_team_user_team" ON "team_user" ("team");

DROP TABLE IF EXISTS "token" CASCADE;
CREATE TABLE "token" (
	"id" TEXT NOT NULL,
	"type" INT NOT NULL,
	"user" TEXT NULL,
	"extra" TEXT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "pk_token" PRIMARY KEY ("id"),
	CONSTRAINT "fk_token_user" FOREIGN KEY ("user") REFERENCES "user" ON DELETE CASCADE
);
CREATE INDEX "idx_token_type_user" ON "token" ("type","user");

DROP TABLE IF EXISTS "vault" CASCADE;
CREATE TABLE "vault" (
	"id" TEXT NOT NULL,
	"team" TEXT NOT NULL,
	"version" INT NOT NULL,
 	"public_key" BYTEA NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "pk_vault" PRIMARY KEY ("team", "id"),
	CONSTRAINT "fk_vault_team" FOREIGN KEY ("team") REFERENCES "team" ON DELETE CASCADE
);

DROP TABLE IF EXISTS "vault_user" CASCADE;
CREATE TABLE "vault_user" (
	"team" TEXT NOT NULL,
	"vault" TEXT NOT NULL,
	"user" TEXT NOT NULL,
	"key" BYTEA NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "pk_vault_user" PRIMARY KEY ("team", "vault", "user"),
	CONSTRAINT "fk_vault" FOREIGN KEY ("team", "vault") REFERENCES "vault" ON DELETE CASCADE,
	CONSTRAINT "fk_vault_user_team_user" FOREIGN KEY ("team", "user") REFERENCES "team_user" ON DELETE CASCADE
);

DROP TABLE IF EXISTS "secret" CASCADE;
CREATE TABLE "secret" (
	"team" TEXT NOT NULL,
	"vault" TEXT NOT NULL,
	"id"  TEXT NOT NULL,
	"version" INT NOT NULL,
	"data" BYTEA NOT NULL,
	"vault_version" INT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "pk_secret" PRIMARY KEY ("team", "vault", "id", "version"),
	CONSTRAINT "fk_secret_team" FOREIGN KEY ("team", "vault" ) REFERENCES "vault" ON DELETE CASCADE
);
CREATE INDEX "idx_secret_team_vault_id" ON "secret" ("team","vault","id");

DROP TABLE IF EXISTS "session" CASCADE;
CREATE TABLE "session" (
	"id" TEXT NOT NULL,
	"user" TEXT NOT NULL,
	"agent" TEXT NOT NULL,
	"requires_csrf" BOOL NOT NULL,
	"last_access" TIMESTAMP WITH TIME ZONE NOT NULL,
	"store_token" TEXT NOT NULL,
	CONSTRAINT "pk_session" PRIMARY KEY ("id"),
	CONSTRAINT "fk_session_user" FOREIGN KEY ("user") REFERENCES "user" ON DELETE CASCADE
);
CREATE INDEX "idx_session_user" ON "session" ("user");
