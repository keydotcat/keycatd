package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/keydotcat/backend/util"
)

type Vault struct {
	Id        string    `scaneo:"pk" json:"id"`
	Team      string    `scaneo:"pk" json:"-"`
	PublicKey []byte    `json:"public_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func createVault(tx *sql.Tx, id, team string, vkp VaultKeyPair) (*Vault, error) {
	v := &Vault{Id: id, Team: team, PublicKey: vkp.PublicKey}
	if err := v.insert(tx); err != nil {
		return nil, err
	}
	for u, k := range vkp.Keys {
		if err := v.addUser(tx, u, k); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (v *Vault) insert(tx *sql.Tx) error {
	if err := v.validate(); err != nil {
		return err
	}
	now := time.Now().UTC()
	v.CreatedAt = now
	v.UpdatedAt = now
	_, err := v.dbInsert(tx)
	switch {
	case isDuplicateErr(err):
		return util.NewErrorFrom(ErrAlreadyExists)
	case isErrOrPanic(err):
		return err
	}
	return nil
}

func (v *Vault) update(tx *sql.Tx) error {
	if err := v.validate(); err != nil {
		return err
	}
	v.UpdatedAt = time.Now().UTC()
	res, err := v.dbUpdate(tx)
	return treatUpdateErr(res, err)
}

func (v Vault) validate() error {
	errs := util.NewErrorFields().(*util.Error)
	if len(v.Id) == 0 {
		errs.SetFieldError("vault_id", "missing")
	}
	if len(v.Team) == 0 {
		errs.SetFieldError("vault_team", "missing")
	}
	if len(v.PublicKey) != publicKeyPackSize {
		errs.SetFieldError("vault_public_key", "invalid")
	}
	return errs.Camo()
}

func (v Vault) AddUsers(ctx context.Context, userKeys map[string][]byte) error {
	for _, k := range userKeys {
		if _, err := verifyAndUnpack(v.PublicKey, k); err != nil {
			return err
		}
	}
	return doTx(ctx, func(tx *sql.Tx) error {
		t := &Team{Id: v.Team}
		users, err := t.getUsers(tx)
		if err != nil {
			return err
		}
		for uk := range userKeys {
			found := false
			for _, user := range users {
				if user.Id == uk {
					found = true
					break
				}
			}
			if !found {
				return util.NewErrorFrom(ErrNotInTeam)
			}
		}
		for u, k := range userKeys {
			if err := v.addUser(tx, u, k); err != nil {
				if isDuplicateErr(err) {
					return util.NewErrorFrom(ErrAlreadyExists)
				}
				return err
			}
		}
		return nil
	})
}

func (v Vault) GetUserIds(ctx context.Context) (uids []string, err error) {
	return uids, doTx(ctx, func(tx *sql.Tx) error {
		uids, err = v.getUserIds(tx)
		return err
	})
}

func (v Vault) getUserIds(tx *sql.Tx) ([]string, error) {
	rows, err := tx.Query(`SELECT "user" FROM "vault_user" WHERE "vault_user"."vault" = $1 AND "vault_user"."team" = $2`, v.Id, v.Team)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	var users []string
	var uid string
	for rows.Next() {
		err = rows.Scan(&uid)
		if isErrOrPanic(err) {
			return nil, util.NewErrorFrom(err)
		}
		users = append(users, uid)
	}
	if err = rows.Err(); isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	return users, nil
}

func (v Vault) addUser(tx *sql.Tx, username string, key []byte) error {
	if err := v.update(tx); err != nil {
		return err
	}
	vu := &vaultUser{Team: v.Team, Vault: v.Id, User: username, Key: key}
	return vu.insert(tx)
}

func (v Vault) RemoveUser(ctx context.Context, username string) error {
	return doTx(ctx, func(tx *sql.Tx) error {
		t := &Team{Id: v.Team}
		tu, err := t.getUserAffiliation(tx, username)
		if err != nil {
			return err
		}
		if tu.Admin {
			return util.NewErrorFrom(err)
		}
		return v.removeUser(tx, username)
	})
}

func (v Vault) removeUser(tx *sql.Tx, username string) error {
	vu := &vaultUser{Team: v.Team, Vault: v.Id, User: username}
	return treatUpdateErr(vu.dbDelete(tx))
}

func (v Vault) AddSecret(ctx context.Context, s *Secret) error {
	s.Team = v.Team
	s.Vault = v.Id
	var err error
	for retry := 0; retry < 3; retry++ {
		s.Id = util.GenerateRandomToken(10)
		err = doTx(ctx, func(tx *sql.Tx) error {
			if err := v.update(tx); err != nil {
				return err
			}
			return s.insert(tx)
		})
		if err == ErrAlreadyExists {
			continue
		}
		break
	}
	return err
}

func (v Vault) UpdateSecret(ctx context.Context, s *Secret) error {
	s.Team = v.Team
	s.Vault = v.Id
	return doTx(ctx, func(tx *sql.Tx) error {
		if err := v.update(tx); err != nil {
			return err
		}
		return s.update(tx)
	})
}

func (v Vault) DeleteSecret(ctx context.Context, sid string) error {
	return doTx(ctx, func(tx *sql.Tx) error {
		if err := v.update(tx); err != nil {
			return err
		}
		res, err := tx.Exec(`DELETE FROM "secret" WHERE "secret"."id" = $1`, sid)
		return treatUpdateErr(res, err)
	})
}

func (v Vault) GetSecrets(ctx context.Context) ([]*Secret, error) {
	db := GetDB(ctx)
	rows, err := db.Query(`SELECT `+selectSecretFields+` FROM "secret" WHERE "secret"."team" = $1 AND "secret"."vault" = $2`, v.Team, v.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	secrets, err := scanSecrets(rows)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	return secrets, nil
}
