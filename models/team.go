package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/keydotcat/backend/util"
)

const DEFAULT_VAULT_NAME = "Generic"

type Team struct {
	Id        string `scaneo:"pk"`
	Name      string
	Owner     string
	Primary   bool
	Size      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func createTeam(tx *sql.Tx, owner *User, primary bool, name string, vaultKeys VaultKeyPair) (*Team, error) {
	now := time.Now().UTC()
	t := &Team{
		util.GenerateRandomToken(16),
		name,
		owner.Id,
		primary,
		0,
		now,
		now,
	}
	if err := t.insert(tx); err != nil {
		return nil, err
	}
	tu := &teamUser{t.Id, owner.Id, true, false}
	if err := tu.insert(tx); err != nil {
		return nil, err
	}
	if err := vaultKeys.checkKeyIdsMatch([]string{owner.Id}); err != nil {
		return nil, err
	}
	if _, err := createVault(tx, DEFAULT_VAULT_NAME, t.Id, vaultKeys); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Team) insert(tx *sql.Tx) error {
	if err := t.validate(); err != nil {
		return err
	}
	_, err := t.dbInsert(tx)
	if err != nil {
		return util.NewErrorf("Could not create team: %s", err)
	}
	return nil
}

func (t *Team) validate() error {
	errs := util.NewErrorFields().(*util.Error)
	if !reValidUsername.MatchString(t.Owner) {
		errs.SetFieldError("username", "invalid")
	}
	if len(t.Name) == 0 {
		errs.SetFieldError("name", "invalid")
	}
	return errs.Camo()
}

func (t *Team) getAdminUsers(tx *sql.Tx) []*User {
	rows, err := tx.Query(`SELECT `+selectUserFullFields+` FROM "user", "team_user" WHERE "team_user"."team" = $1 AND "user"."id" = "team_user"."user" AND "team_user"."admin" = true`, t.Id)
	if err != nil {
		panic("Could not retrieve users: " + err.Error())
	}
	users, err := scanUsers(rows)
	if err != nil {
		panic("Could not scan users: " + err.Error())
	}
	return users
}

func (t *Team) getUsers(tx *sql.Tx) []*User {
	rows, err := tx.Query(`SELECT `+selectUserFullFields+` FROM "user", "team_user" WHERE "team_user"."team" = $1 AND "user"."id" = "team_user"."user"`, t.Id)
	if err != nil {
		panic("Could not retrieve users: " + err.Error())
	}
	users, err := scanUsers(rows)
	if err != nil {
		panic("Could not scan users: " + err.Error())
	}
	return users
}

func (t *Team) getTeamUsers(tx *sql.Tx) []*teamUser {
	rows, err := tx.Query(`SELECT `+selectTeamUserFullFields+` FROM "team_user" WHERE "team_user"."team" = $1`, t.Id)
	if err != nil {
		panic("Could not retrieve team user states: " + err.Error())
	}
	users, err := scanTeamUsers(rows)
	if err != nil {
		panic("Could not scan team users: " + err.Error())
	}
	return users
}

func (t *Team) CreateVault(ctx context.Context, name string, vaultKeys VaultKeyPair) (v *Vault, err error) {
	return v, doTx(ctx, func(tx *sql.Tx) error {
		v, err = createVault(tx, name, t.Id, vaultKeys)
		return err
	})
}

func (t *Team) filterTeamUsers(tx *sql.Tx, uids ...string) ([]*teamUser, error) {
	teamUsers := make([]*teamUser, len(uids))
	tus := t.getTeamUsers(tx)
	for i, uid := range uids {
		belongsToTeam := false
		for _, tu := range tus {
			if tu.User == uid {
				teamUsers[i] = tu
				belongsToTeam = true
				break
			}
		}
		if !belongsToTeam {
			return nil, util.NewErrorf("%s does not belong to team", uid)
		}
	}
	return teamUsers, nil
}

func (t *Team) PromoteUser(ctx context.Context, promoter *User, promotee *User, vaultKeys VaultKeyPair) error {
	return doTx(ctx, func(tx *sql.Tx) error {
		teamUsers, err := t.filterTeamUsers(tx, promoter.Id, promotee.Id)
		if err != nil {
			return err
		}
		if !teamUsers[0].Admin {
			return util.NewErrorf("You're not an admin!")
		}
		if teamUsers[1].Admin {
			return util.NewErrorf("%s is admin", promotee.Id)
		}
		missingVaults := t.getVaultsMissingForUser(tx, promotee)
		vaultIds := make([]string, len(missingVaults))
		for i, v := range missingVaults {
			vaultIds[i] = v.Id
		}
		if err := vaultKeys.checkKeyIdsMatch(vaultIds); err != nil {
			return err
		}
		for _, v := range missingVaults {
			if err := v.addUser(tx, promotee.Id, vaultKeys.Keys[v.Id]); err != nil {
				return err
			}
		}
		ta := teamUsers[1]
		ta.Admin = true
		return ta.update(tx)
	})
}

func (t *Team) demoteUser(ctx context.Context, demoter *User, demotee *User) error {
	return doTx(ctx, func(tx *sql.Tx) error {
		teamUsers, err := t.filterTeamUsers(tx, demoter.Id, demotee.Id)
		if err != nil {
			return err
		}
		if !teamUsers[0].Admin {
			return util.NewErrorf("You're not an admin!")
		}
		if !teamUsers[1].Admin {
			return util.NewErrorf("%s is not an admin", demotee.Id)
		}
		ta := teamUsers[1]
		ta.Admin = false
		return ta.update(tx)
	})
}

func (t *Team) GetVaultsForUser(ctx context.Context, u *User) []*Vault {
	db := GetDB(ctx)
	rows, err := db.Query(`SELECT `+selectVaultFullFields+` FROM "vault", "vault_user" WHERE  "vault"."team" = $1 AND "vault"."team" = "vault_user"."team" AND "vault"."id" = "vault_user"."vault" AND "vault_user"."user" = $2`, t.Id, u.Id)
	if err != nil {
		panic("Could not retrieve vaults: " + err.Error())
	}
	vaults, err := scanVaults(rows)
	if err != nil {
		panic("Could not scan vaults: " + err.Error())
	}
	return vaults
}

func (t *Team) getVaultsMissingForUser(tx *sql.Tx, u *User) []*Vault {
	rows, err := tx.Query(`SELECT `+selectVaultFullFields+` FROM "vault" WHERE "vault"."id" NOT IN ( SELECT "vault"."id" FROM "vault", "vault_user" WHERE "vault"."team" = $1 AND "vault"."team" = "vault_user"."team" AND "vault"."id" = "vault_user"."vault" AND "vault_user"."user" = $2`, t.Id, u.Id)
	if err != nil {
		panic("Could not retrieve vaults: " + err.Error())
	}
	vaults, err := scanVaults(rows)
	if err != nil {
		panic("Could not scan vaults: " + err.Error())
	}
	return vaults
}
