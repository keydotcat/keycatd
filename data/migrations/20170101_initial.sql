DROP TABLE IF EXISTS "Users";
CREATE TABLE "Users" (
	"Id" STRING NOT NULL,
	"Email" STRING NOT NULL,
	"UnconfirmedEmail" STRING NULL,
	"HashPass" BYTEA NOT NULL,
	"FullName" STRING NOT NULL,
	"ConfirmedAt" TIMESTAMP WITH TIME ZONE NULL,
	"LockedAt" TIMESTAMP WITH TIME ZONE NULL,
	"SignInCount" INT NULL DEFAULT 0,
	"FailedAttempts" INT NULL DEFAULT 0,
	"PublicKey" BYTEA NOT NULL,
	"Key" BYTEA NOT NULL,
	"CreatedAt" TIMESTAMP WITH TIME ZONE,
	"UpdatedAt" TIMESTAMP WITH TIME ZONE,
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	FAMILY "primary" ("Id", "Email", "UnconfirmedEmail", "HashPass", "FullName", "ConfirmedAt", "LockedAt", "SignInCount", "FailedAttempts", "PublicKey", "Key", "CreatedAt", "UpdatedAt")
);

DROP TABLE IF EXISTS "Teams";
CREATE TABLE "Teams" (
	"Id" STRING NOT NULL,
	"Name" STRING NOT NULL,
	"Owner" STRING NOT NULL REFERENCES "Users" ("Id"),
	"BelongsToUser" BOOL NOT NULL,
	"Size" INT NOT NULL,
	"CreatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	"UpdatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	FAMILY "primary" ("Id", "Owner", "Size", "CreatedAt", "UpdatedAt")
);

DROP TABLE IF EXISTS "Invites";
CREATE TABLE "Invites" (
	"Team" STRING NOT NULL REFERENCES "Teams" ("Id"),
	"Email" STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Email"),
	INDEX email ("Email" ASC),
	FAMILY "primary" ("Team", "Email")
) INTERLEAVE IN PARENT "Teams" ("Team");

DROP TABLE IF EXISTS "Team_Users";
CREATE TABLE "Team_Users" (
	"Team" STRING NOT NULL REFERENCES "Teams" ("Id"),
	"Username" STRING NOT NULL REFERENCES "Users" ("Id"),
	"Admin" BOOL NOT NULL,
	"RequiresAccess" BOOL NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Username"),
	INDEX "team" ("Team"),
	FAMILY "primary" ("Team", "Username", "Admin", "RequiresAccess")
) INTERLEAVE IN PARENT "Teams" ("Team");

DROP TABLE IF EXISTS "Tokens";
CREATE TABLE "Tokens" (
	"Id" STRING NOT NULL,
	"Type" INT NOT NULL,
	"Username" STRING NULL REFERENCES "Users" ("Id"),
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	FAMILY "primary" ("Id", "Type", "Username")
);

DROP TABLE IF EXISTS "Vaults";
CREATE TABLE "Vaults" (
	"Id" STRING NOT NULL,
	"Team" STRING NOT NULL REFERENCES "Teams" ("Id"),
 	"PublicKey" STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Id" ASC),
	FAMILY "primary" ("Id", "Team", "PublicKey")
) INTERLEAVE IN PARENT "Teams" ("Team");

DROP TABLE IF EXISTS "Vault_User";
CREATE TABLE "Vault_User" (
	"Team" STRING NOT NULL, 
	"Vault" STRING NOT NULL,
	"Username" STRING NOT NULL,
	"Key" STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Vault", "Username" ASC),
	FAMILY "primary" ("Team", "Vault", "Username", "Key"),
	CONSTRAINT "fk_TeamsVaults" FOREIGN KEY ("Team", "Vault" ) REFERENCES "Vaults"
) INTERLEAVE IN PARENT "Vaults" ("Team", "Vault");

DROP TABLE IF EXISTS "Secrets";
CREATE TABLE "Secrets" (
	"Team" STRING NOT NULL,
	"Vault" STRING NOT NULL,
	"Id"  STRING NOT NULL,
	"Data" BYTES NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Vault", "Id" ASC),
	FAMILY "primary" ("Team", "Vault", "Id",  "Data"),
	CONSTRAINT "fk_TeamsVaults" FOREIGN KEY ("Team", "Vault" ) REFERENCES "Vaults"
) INTERLEAVE IN PARENT "Vaults" ("Team","Vault");

