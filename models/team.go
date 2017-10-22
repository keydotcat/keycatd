package models

import (
	"context"
	"database/sql"
	"fmt"
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
	isErrOrPanic(err)
	return util.NewErrorFrom(err)
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

func (t *Team) getAdminUsers(tx *sql.Tx) ([]*User, error) {
	rows, err := tx.Query(`SELECT `+selectUserFullFields+` FROM "user", "team_user" WHERE "team_user"."team" = $1 AND "user"."id" = "team_user"."user" AND "team_user"."admin" = true`, t.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	users, err := scanUsers(rows)
	isErrOrPanic(err)
	return users, util.NewErrorFrom(err)
}

func (t *Team) getUsers(tx *sql.Tx) ([]*User, error) {
	rows, err := tx.Query(`SELECT `+selectUserFullFields+` FROM "user", "team_user" WHERE "team_user"."team" = $1 AND "user"."id" = "team_user"."user"`, t.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	users, err := scanUsers(rows)
	isErrOrPanic(err)
	return users, util.NewErrorFrom(err)
}

func (t *Team) getUsersAfiliation(tx *sql.Tx) ([]*teamUser, error) {
	rows, err := tx.Query(`SELECT `+selectTeamUserFullFields+` FROM "team_user" WHERE "team_user"."team" = $1`, t.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	users, err := scanTeamUsers(rows)
	isErrOrPanic(err)
	return users, util.NewErrorFrom(err)
}

func (t *Team) CreateVault(ctx context.Context, name string, vaultKeys VaultKeyPair) (v *Vault, err error) {
	return v, doTx(ctx, func(tx *sql.Tx) error {
		v, err = createVault(tx, name, t.Id, vaultKeys)
		return err
	})
}

func (t *Team) filterTeamUsers(tx *sql.Tx, uids ...string) ([]*teamUser, error) {
	binds := make([]string, len(uids)+1)
	binds[0] = t.Id
	for i := range uids {
		//Pos x of array is $(x+1) since binds start at index 1
		binds[i+1] = fmt.Sprintf("$%d", i+2)
	}
	stmt := fmt.Sprintf(`SELECT %s FROM "team_user" WHERE "team_user"."team" = $1 AND "team_user"."user" in (%s)`, selectTeamUserFields, binds)
	rows, err := tx.Query(stmt, binds)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	tus, err := scanTeamUsers(rows)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	teamUsers := make([]*teamUser, len(uids))
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
			return nil, util.NewErrorFrom(ErrNotInTeam)
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
			return util.NewErrorFrom(ErrUnauthorized)
		}
		if teamUsers[1].Admin {
			return nil
		}
		missingVaults, err := t.getVaultsMissingForUser(tx, promotee)
		if err != nil {
			return err
		}
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

func (t *Team) AddOrInviteUserByEmail(ctx context.Context, admin *User, newcomerEmail string) (bool, error) {
	add := false
	return add, doTx(ctx, func(tx *sql.Tx) error {
		nu, err := findUserByEmail(tx, newcomerEmail)
		if err != nil {
			return err
		}
		if nu != nil {
			add = true
			return t.addUser(tx, admin, nu)
		}
		return t.generateInvite(tx, admin, newcomerEmail)
	})
}

func (t *Team) getUserAffiliation(tx *sql.Tx, u *User) (*teamUser, error) {
	tu := &teamUser{Team: t.Id, User: u.Id}
	err := tu.dbFind(tx)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case isErrOrPanic(err):
		return nil, util.NewErrorFrom(err)
	default:
		return tu, nil
	}
}

func (t *Team) generateInvite(tx *sql.Tx, admin *User, email string) error {
	if !reValidEmail.MatchString(email) {
		return util.NewErrorFrom(ErrInvalidEmail)
	}
	if err := t.checkAdmin(tx, admin); err != nil {
		return err
	}
	i := &Invite{Team: t.Id, Email: email}
	return i.insert(tx)
}

func (t *Team) checkAdmin(tx *sql.Tx, u *User) error {
	tu, err := t.getUserAffiliation(tx, u)
	if err != nil {
		return err
	}
	if tu == nil {
		return util.NewErrorFrom(ErrNotInTeam)
	} else if !tu.Admin {
		return util.NewErrorFrom(ErrUnauthorized)
	}
	return nil
}

func (t *Team) addUser(tx *sql.Tx, admin *User, newUser *User) error {
	if err := t.checkAdmin(tx, admin); err != nil {
		return err
	}
	tu, err := t.getUserAffiliation(tx, newUser)
	if err != nil {
		return err
	}
	if tu != nil {
		return util.NewErrorFrom(ErrAlreadyInTeam)
	}
	tu = &teamUser{t.Id, newUser.Id, false, false}
	return tu.insert(tx)
}

func (t *Team) demoteUser(ctx context.Context, demoter *User, demotee *User) error {
	return doTx(ctx, func(tx *sql.Tx) error {
		teamUsers, err := t.filterTeamUsers(tx, demoter.Id, demotee.Id)
		if err != nil {
			return err
		}
		if !teamUsers[0].Admin {
			return util.NewErrorFrom(ErrUnauthorized)
		}
		if !teamUsers[1].Admin {
			return nil
		}
		ta := teamUsers[1]
		ta.Admin = false
		return ta.update(tx)
	})
}

func (t *Team) GetVaultsForUser(ctx context.Context, u *User) ([]*Vault, error) {
	db := GetDB(ctx)
	rows, err := db.Query(`SELECT `+selectVaultFullFields+` FROM "vault", "vault_user" WHERE  "vault"."team" = $1 AND "vault"."team" = "vault_user"."team" AND "vault"."id" = "vault_user"."vault" AND "vault_user"."user" = $2`, t.Id, u.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	vaults, err := scanVaults(rows)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	return vaults, nil
}

func (t *Team) getVaultsMissingForUser(tx *sql.Tx, u *User) ([]*Vault, error) {
	rows, err := tx.Query(`SELECT `+selectVaultFullFields+` FROM "vault" WHERE "vault"."id" NOT IN ( SELECT "vault"."id" FROM "vault", "vault_user" WHERE "vault"."team" = $1 AND "vault"."team" = "vault_user"."team" AND "vault"."id" = "vault_user"."vault" AND "vault_user"."user" = $2`, t.Id, u.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	vaults, err := scanVaults(rows)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	return vaults, nil
}
