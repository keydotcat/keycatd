DROP TABLE IF EXISTS "User";
CREATE TABLE "User" (
	"Id" STRING NOT NULL,
	"Email" STRING NOT NULL,
	"UnconfirmedEmail" STRING NULL,
	"Password" STRING NOT NULL,
	"FullName" STRING NOT NULL,
	"ConfirmedAt" TIMESTAMP WITH TIME ZONE NULL,
	"LockedAt" TIMESTAMP WITH TIME ZONE NULL,
	"SignInCount" INT NULL DEFAULT 0,
	"FailedAttempts" INT NULL DEFAULT 0,
	"PublicKey" STRING NOT NULL,
	"Key" STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	FAMILY "primary" ("Id", "Email", "UnconfirmedEmail", "Password", "FullName", "ConfirmedAt", "LockedAt", "SignInCount", "FailedAttempts", "PublicKey", "Key")
);

DROP TABLE IF EXISTS "Invites";
CREATE TABLE "Invites" (
	"Team" STRING NOT NULL,
	"Email" STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team" ASC),
	INDEX email ("Email" ASC),
	FAMILY "primary" ("Team", "Email")
);

DROP TABLE IF EXISTS "Team";
CREATE TABLE "Team" (
	"Id" STRING NOT NULL,
	"Owner" STRING NOT NULL,
	"Size" INT NOT NULL,
	"CreatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	"UpdatedAt" TIMESTAMP WITH TIME ZONE NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	FAMILY "primary" ("Id", "Owner", "Size", "CreatedAt", "UpdatedAt")
);

DROP TABLE IF EXISTS "Team_Users";
CREATE TABLE "Team_Users" (
	"Team" STRING NOT NULL,
	"Username" STRING NOT NULL,
	"Admin" BOOL NOT NULL,
	"RequiresAccess" BOOL NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Team", "Username" ASC),
	FAMILY "primary" ("Team", "Username", "Admin", "RequiresAccess")
);

DROP TABLE IF EXISTS "Token";
CREATE TABLE "Token" (
	"Id" STRING NOT NULL,
	type INT NOT NULL,
	username STRING NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	CONSTRAINT username FOREIGN KEY (username) REFERENCES "User" ("Id"),
	FAMILY "primary" ("Id", type, username)
);

DROP TABLE IF EXISTS "Vault";
CREATE TABLE "Vault" (
	"Id" STRING NOT NULL,
	"Team" STRING NOT NULL,
	"PublicKey" STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	CONSTRAINT team FOREIGN KEY ("Team") REFERENCES "Team" ("Id"),
	FAMILY "primary" ("Id", "Team", "PublicKey")
);

DROP TABLE IF EXISTS "Vault_User";
CREATE TABLE "Vault_User" (
	"Vault" STRING NOT NULL,
	"Username" STRING NOT NULL,
	"Key" STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Vault", "Username" ASC),
	FAMILY "primary" ("Vault", "Username", "Key")
);

DROP TABLE IF EXISTS "Secret";
CREATE TABLE "Secret" (
	"Id" INT NOT NULL DEFAULT unique_rowid(),
	"Vault" STRING NULL,
	"Data" BYTES NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY ("Id" ASC),
	CONSTRAINT "fk_Vault_ref_Vault" FOREIGN KEY ("Vault") REFERENCES "Vault" ("Id"),
	FAMILY "primary" ("Id", "Vault", "Data")
);


