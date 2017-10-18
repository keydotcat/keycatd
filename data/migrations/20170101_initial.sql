DROP TABLE IF EXISTS "User";
CREATE TABLE "User" (
	"Id" STRING NOT NULL,
	"Email" STRING NOT NULL,
	"UnconfirmedEmail" STRING NULL,
	"HashPass" BLOB NOT NULL,
	"FullName" STRING NOT NULL,
	"ConfirmedAt" TIMESTAMP WITH TIME ZONE NULL,
	"LockedAt" TIMESTAMP WITH TIME ZONE NULL,
	"SignInCount" INT NULL DEFAULT 0,
	"FailedAttempts" INT NULL DEFAULT 0,
	"PublicKey" BLOB NOT NULL,
	"Key" BLOB NOT NULL,
	"CreatedAt" TIMESTAMP WITH TIME ZONE,
	"UpdatedAt" TIMESTAMP WITH TIME ZONE,
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	FAMILY "primary" ("Id", "Email", "UnconfirmedEmail", "HashPass", "FullName", "ConfirmedAt", "LockedAt", "SignInCount", "FailedAttempts", "PublicKey", "Key", "CreatedAt", "UpdatedAt")
);

DROP TABLE IF EXISTS "Team";
CREATE TABLE "Team" (
	"Id" STRING NOT NULL,
	"Name" STRING NOT NULL,
	"Owner" STRING NOT NULL REFERENCES "User" ("Id"),
	"BelongsToUser" BOOL NOT NULL,
	"Size" INT NOT NULL,
	"CreatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	"UpdatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	FAMILY "primary" ("Id", "Owner", "Size", "CreatedAt", "UpdatedAt")
);

DROP TABLE IF EXISTS "Invites";
CREATE TABLE "Invites" (
	"Team" STRING NOT NULL REFERENCES "Team" ("Id"),
	"Email" STRING NOT NULL,
	"CreatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Email"),
	INDEX email ("Email" ASC),
	FAMILY "primary" ("Team", "Email", "CreatedAt")
) INTERLEAVE IN PARENT "Team" ("Team");

DROP TABLE IF EXISTS "Team_Users";
CREATE TABLE "Team_Users" (
	"Team" STRING NOT NULL REFERENCES "Team" ("Id"),
	"Username" STRING NOT NULL REFERENCES "User" ("Id"),
	"Admin" BOOL NOT NULL,
	"RequiresAccess" BOOL NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Username"),
	INDEX "team" ("Team"),
	FAMILY "primary" ("Team", "Username", "Admin", "RequiresAccess")
) INTERLEAVE IN PARENT "Team" ("Team");

DROP TABLE IF EXISTS "Tokens";
CREATE TABLE "Tokens" (
	"Id" STRING NOT NULL,
	"Type" INT NOT NULL,
	"Username" STRING NULL REFERENCES "User" ("Id"),
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	FAMILY "primary" ("Id", "Type", "Username")
);

DROP TABLE IF EXISTS "Vault";
CREATE TABLE "Vault" (
	"Id" STRING NOT NULL,
	"Team" STRING NOT NULL REFERENCES "Team" ("Id"),
 	"PublicKey" BLOB NOT NULL,
	"CreatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	"UpdatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Id" ASC),
	FAMILY "primary" ("Id", "Team", "PublicKey", "CreatedAt", "UpdatedAt")
) INTERLEAVE IN PARENT "Team" ("Team");

DROP TABLE IF EXISTS "Vault_User";
CREATE TABLE "Vault_User" (
	"Team" STRING NOT NULL, 
	"Vault" STRING NOT NULL,
	"Username" STRING NOT NULL,
	"Key" STRING NOT NULL,
	"CreatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	"UpdatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Vault", "Username" ASC),
	FAMILY "primary" ("Team", "Vault", "Username", "Key", "CreatedAt", "UpdatedAt"),
	CONSTRAINT "fk_Team" FOREIGN KEY ("Team", "Vault" ) REFERENCES "Vault"
) INTERLEAVE IN PARENT "Vault" ("Team", "Vault");

DROP TABLE IF EXISTS "Secret";
CREATE TABLE "Secret" (
	"Team" STRING NOT NULL,
	"Vault" STRING NOT NULL,
	"Id"  STRING NOT NULL,
	"Data" BYTES NOT NULL,
	"CreatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	"UpdatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Vault", "Id" ASC),
	FAMILY "primary" ("Team", "Vault", "Id",  "Data", "UpdatedAt", "CreatedAt"),
	CONSTRAINT "fk_Team" FOREIGN KEY ("Team", "Vault" ) REFERENCES "Vault"
) INTERLEAVE IN PARENT "Vault" ("Team","Vault");

