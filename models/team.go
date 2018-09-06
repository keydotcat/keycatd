package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/keydotcat/backend/util"
)

const DEFAULT_VAULT_NAME = "Generic"

type Team struct {
	Id        string    `scaneo:"pk" json:"id"`
	Name      string    `json:"name"`
	Owner     string    `json:"owner"`
	Primary   bool      `json:"primary"`
	Size      int       `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
		errs.SetFieldError("team_owner", "invalid")
	}
	if len(t.Name) == 0 {
		errs.SetFieldError("team_name", "invalid")
	}
	return errs.SetErrorOrCamo(ErrInvalidAttributes)
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

func (t *Team) CreateVault(ctx context.Context, u *User, name string, signedVaultKeys VaultKeyPair) (v *Vault, err error) {
	vaultKeys, err := signedVaultKeys.verifyAndUnpack(u.PublicKey)
	if err != nil {
		return nil, err
	}
	return v, doTx(ctx, func(tx *sql.Tx) error {
		admins, err := t.getAdminUsers(tx)
		if err != nil {
			return err
		}
		isAdmin := false
		uids := make([]string, len(admins))
		for i, admin := range admins {
			uids[i] = admin.Id
			isAdmin = isAdmin || (admin.Id == u.Id)
		}
		if !isAdmin {
			return util.NewErrorFrom(ErrUnauthorized)
		}
		if err = vaultKeys.checkKeyIdsMatch(uids); err != nil {
			return err
		}
		v, err = createVault(tx, name, t.Id, vaultKeys)
		return err
	})
}

func (t *Team) filterTeamUsers(tx *sql.Tx, uids ...string) ([]*teamUser, error) {
	bindValues := make([]interface{}, len(uids)+1)
	bindIds := make([]string, len(uids))
	bindValues[0] = t.Id
	for i := range uids {
		//Pos x of array is $(x+1) since binds start at index 1
		bindIds[i] = fmt.Sprintf("$%d", i+2)
		bindValues[i+1] = uids[i]
	}
	stmt := fmt.Sprintf(`SELECT %s FROM "team_user" WHERE "team_user"."team" = $1 AND "team_user"."user" in (%s)`, selectTeamUserFields, strings.Join(bindIds, ","))
	rows, err := tx.Query(stmt, bindValues...)
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

func (t *Team) PromoteUser(ctx context.Context, promoter *User, promotee *User, signedVaultKeys VaultKeyPair) error {
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
		if err := signedVaultKeys.checkKeyIdsMatch(vaultIds); err != nil {
			return err
		}
		for _, v := range missingVaults {
			vaultKey := signedVaultKeys.Keys[v.Id]
			_, err := verifyAndUnpack(v.PublicKey, vaultKey)
			if err != nil {
				return err
			}
			if err := v.addUser(tx, promotee.Id, vaultKey); err != nil {
				return err
			}
		}
		ta := teamUsers[1]
		ta.Admin = true
		return ta.update(tx)
	})
}

func (t *Team) AddOrInviteUserByEmail(ctx context.Context, admin *User, newcomerEmail string) (i *Invite, err error) {
	return i, doTx(ctx, func(tx *sql.Tx) error {
		nu, err := findUserByEmail(tx, newcomerEmail)
		switch {
		case util.CheckErr(err, ErrDoesntExist):
			i, err = t.generateInvite(tx, admin, newcomerEmail)
			return err
		case err != nil:
			return err
		default:
			return t.addUser(tx, admin, nu)
		}
	})
}

func (t *Team) getUserAffiliation(tx *sql.Tx, username string) (*teamUser, error) {
	tu := &teamUser{Team: t.Id, User: username}
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

func (t *Team) generateInvite(tx *sql.Tx, admin *User, email string) (*Invite, error) {
	if !reValidEmail.MatchString(email) {
		return nil, util.NewErrorFrom(ErrInvalidEmail)
	}
	if err := t.checkAdmin(tx, admin); err != nil {
		return nil, err
	}
	i := &Invite{Team: t.Id, Email: email}
	return i, i.insert(tx)
}

func (t *Team) getInvites(tx *sql.Tx) ([]*Invite, error) {
	rows, err := tx.Query(`SELECT `+selectInviteFields+` FROM "invite" WHERE "invite"."team" = $1`, t.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	invites, err := scanInvites(rows)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	return invites, nil
}

func (t *Team) CheckAdmin(ctx context.Context, u *User) (isAdmin bool, err error) {
	return isAdmin, doTx(ctx, func(tx *sql.Tx) error {
		err = t.checkAdmin(tx, u)
		switch {
		case err == nil:
			isAdmin = true
		case util.CheckErr(err, ErrUnauthorized):
			isAdmin = false
			err = nil
		}
		return err
	})
}

func (t *Team) checkAdmin(tx *sql.Tx, u *User) error {
	tu, err := t.getUserAffiliation(tx, u.Id)
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
	return t.addUserNoAdminCheck(tx, newUser)
}

func (t *Team) addUserNoAdminCheck(tx *sql.Tx, newUser *User) error {
	tu, err := t.getUserAffiliation(tx, newUser.Id)
	if err != nil {
		return err
	}
	if tu != nil {
		return util.NewErrorFrom(ErrAlreadyInTeam)
	}
	tu = &teamUser{t.Id, newUser.Id, false, false}
	return tu.insert(tx)
}

func (t *Team) DemoteUser(ctx context.Context, demoter *User, demotee *User) error {
	if t.Owner == demotee.Id {
		return util.NewErrorFrom(ErrUnauthorized)
	}
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

func (t *Team) GetSecretsForUser(ctx context.Context, u *User) (s []*Secret, err error) {
	return s, doTx(ctx, func(tx *sql.Tx) error {
		s, err = t.getSecretsForUser(tx, u)
		return err
	})
}

func (t *Team) getSecretsForUser(tx *sql.Tx, u *User) (s []*Secret, err error) {
	query := `
	SELECT DISTINCT ON ("secret"."team", "secret"."vault", "secret"."id")
		"secret"."team", "secret"."vault", "secret"."id", "secret"."version", "secret"."data", "secret"."vault_version", "secret"."created_at"  
	FROM "secret", "vault_user" 
	WHERE 
		"secret"."team" = $1 AND 
		"secret"."team" = "vault_user"."team" AND 
		"secret"."vault" = "vault_user"."vault" AND 
		"vault_user"."user" = $2
	ORDER BY "secret"."team", "secret"."vault", "secret"."id", "secret"."version" DESC`
	rows, err := tx.Query(query, t.Id, u.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	secrets, err := scanSecrets(rows)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	return secrets, nil
}

func (t *Team) GetVaultForUser(ctx context.Context, vid string, u *User) (*Vault, error) {
	db := GetDB(ctx)
	r := db.QueryRow(`SELECT `+selectVaultFullFields+` FROM "vault", "vault_user" WHERE "vault"."team" = $1 AND "vault"."id" = $2 AND "vault_user"."team" = "vault"."team" AND "vault_user"."user" = $3 AND "vault_user"."vault" = "vault"."id"`, t.Id, vid, u.Id)
	v := &Vault{}
	err := v.dbScanRow(r)
	if isNotExistsErr(err) {
		return nil, util.NewErrorFrom(ErrDoesntExist)
	}
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	return v, nil
}

func (t *Team) GetVaultsForUser(ctx context.Context, u *User) (vs []*Vault, err error) {
	return vs, doTx(ctx, func(tx *sql.Tx) error {
		vs, err = t.getVaultsForUser(tx, u)
		return err
	})
}

func (t *Team) getVaultsForUser(tx *sql.Tx, u *User) ([]*Vault, error) {
	rows, err := tx.Query(`SELECT `+selectVaultFullFields+` FROM "vault", "vault_user" WHERE  "vault"."team" = $1 AND "vault"."team" = "vault_user"."team" AND "vault"."id" = "vault_user"."vault" AND "vault_user"."user" = $2`, t.Id, u.Id)
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
	cmd := `SELECT ` + selectVaultFullFields + ` FROM "vault" WHERE "vault"."team" = $1 AND "vault"."id" NOT IN ( SELECT "vault"."id" FROM "vault", "vault_user" WHERE "vault"."team" = $1 AND "vault"."team" = "vault_user"."team" AND "vault"."id" = "vault_user"."vault" AND "vault_user"."user" = $2)`
	rows, err := tx.Query(cmd, t.Id, u.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	vaults, err := scanVaults(rows)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	return vaults, nil
}
