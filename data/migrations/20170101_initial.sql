DROP TABLE IF EXISTS "user";
CREATE TABLE "user" (
	"id" STRING NOT NULL,
	"email" STRING NOT NULL,
	"unconfirmed_email" STRING NULL,
	"hash_pass" BLOB NOT NULL,
	"full_name" STRING NOT NULL,
	"confirmed_at" TIMESTAMP WITH TIME ZONE NULL,
	"locked_at" TIMESTAMP WITH TIME ZONE NULL,
	"sign_in_count" INT NULL DEFAULT 0,
	"failed_attempts" INT NULL DEFAULT 0,
	"public_key" BLOB NOT NULL,
	"key" BLOB NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE,
	"updated_at" TIMESTAMP WITH TIME ZONE,
	CONSTRAINT "primary" PRIMARY KEY ("id" ASC),
	FAMILY "primary" ("id", "email", "unconfirmed_email", "hash_pass", "full_name", "confirmed_at", "locked_at", "sign_in_count", "failed_attempts", "public_key", "key", "created_at", "updated_at")
);

DROP TABLE IF EXISTS "team";
CREATE TABLE "team" (
	"id" STRING NOT NULL,
	"name" STRING NOT NULL,
	"owner" STRING NOT NULL REFERENCES "user" ("id"),
	"primary" BOOL NOT NULL,
	"size" INT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("id" ASC),
	FAMILY "primary" ("id", "name", "owner", "primary", "size", "created_at", "updated_at")
);

DROP TABLE IF EXISTS "invite";
CREATE TABLE "invite" (
	"team" STRING NOT NULL REFERENCES "team" ("id"),
	"email" STRING NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("team", "email"),
	INDEX "idx_email" ("email" ASC),
	FAMILY "primary" ("team", "email", "created_at")
) INTERLEAVE IN PARENT "team" ("team");

DROP TABLE IF EXISTS "team_user";
CREATE TABLE "team_user" (
	"team" STRING NOT NULL REFERENCES "team" ("id"),
	"user" STRING NOT NULL REFERENCES "user" ("id"),
	"admin" BOOL NOT NULL,
	"access_required" BOOL NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("team", "user"),
	INDEX "idx_team" ("team"),
	FAMILY "primary" ("team", "user", "admin", "access_required")
) INTERLEAVE IN PARENT "team" ("team");

DROP TABLE IF EXISTS "token";
CREATE TABLE "token" (
	"id" STRING NOT NULL,
	"type" INT NOT NULL,
	"user" STRING NULL REFERENCES "user" ("id"),
	"extra" STRING NULL, 
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("id" ASC),
	INDEX "idx_type_user" ("type","user"),
	FAMILY "primary" ("id", "type", "user", "extra", "created_at", "updated_at")
);

DROP TABLE IF EXISTS "vault";
CREATE TABLE "vault" (
	"id" STRING NOT NULL,
	"team" STRING NOT NULL REFERENCES "team" ("id"),
 	"public_key" BLOB NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("team", "id" ASC),
	FAMILY "primary" ("id", "team", "public_key", "created_at", "updated_at")
) INTERLEAVE IN PARENT "team" ("team");

DROP TABLE IF EXISTS "vault_user";
CREATE TABLE "vault_user" (
	"team" STRING NOT NULL, 
	"vault" STRING NOT NULL,
	"user" STRING NOT NULL,
	"key" BLOB NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("team", "vault", "user" ASC),
	FAMILY "primary" ("team", "vault", "user", "key", "created_at", "updated_at"),
	CONSTRAINT "fk_Vault" FOREIGN KEY ("team", "vault") REFERENCES "vault",
	CONSTRAINT "fk_TeamUser" FOREIGN KEY ("team", "user") REFERENCES "team_user"
) INTERLEAVE IN PARENT "vault" ("team", "vault");

DROP TABLE IF EXISTS "secret";
CREATE TABLE "secret" (
	"team" STRING NOT NULL,
	"vault" STRING NOT NULL,
	"id"  STRING NOT NULL,
	"data" BLOB NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("team", "vault", "id" ASC),
	FAMILY "primary" ("team", "vault", "id",  "data", "created_at", "updated_at"),
	CONSTRAINT "fk_Team" FOREIGN KEY ("team", "vault" ) REFERENCES "vault"
) INTERLEAVE IN PARENT "vault" ("team","vault");

