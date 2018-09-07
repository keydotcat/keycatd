package models

import (
	"database/sql"
	"time"

	"github.com/keydotcat/server/util"
)

type vaultUser struct {
	Team      string `scaneo:"pk"`
	Vault     string `scaneo:"pk"`
	User      string `scaneo:"pk"`
	Key       []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tu *vaultUser) insert(tx *sql.Tx) error {
	if err := tu.validate(); err != nil {
		return err
	}
	now := time.Now().UTC()
	tu.CreatedAt = now
	tu.UpdatedAt = now
	_, err := tu.dbInsert(tx)
	if IsDuplicateErr(err) {
		return util.NewErrorf("User %s is already in vault", tu.User)
	}
	if isErrOrPanic(err) {
		return util.NewErrorFrom(err)
	}
	return nil
}

func (v vaultUser) validate() error {
	errs := util.NewErrorFields().(*util.Error)
	if len(v.Team) == 0 {
		errs.SetFieldError("vaultuser_team", "missing")
	}
	if len(v.Vault) == 0 {
		errs.SetFieldError("vaultuser_vault", "missing")
	}
	if len(v.User) == 0 {
		errs.SetFieldError("vaultuser_user", "missing")
	}
	if len(v.Key) != privateKeyPackSize {
		errs.SetFieldError("vaultuser_key", "invalid")
	}
	return errs.SetErrorOrCamo(ErrInvalidAttributes)
}
