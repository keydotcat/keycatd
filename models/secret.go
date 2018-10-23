package models

import (
	"database/sql"
	"time"

	"github.com/keydotcat/keycatd/util"
)

type Secret struct {
	Team         string    `scaneo:"pk" json:"-"`
	Vault        string    `scaneo:"pk" json:"vault"`
	Id           string    `scaneo:"pk" json:"id"`
	Version      uint32    `json:"version"`
	Data         []byte    `json:"data"`
	VaultVersion uint32    `json:"vault_version"`
	CreatedAt    time.Time `json:"created_at"`
}

func (v *Secret) insert(tx *sql.Tx) error {
	v.Id = util.GenerateRandomToken(10)
	v.Version = 1
	v.CreatedAt = time.Now().UTC()
	if err := v.validate(false); err != nil {
		return err
	}
	_, err := v.dbInsert(tx)
	switch {
	case IsDuplicateErr(err):
		return util.NewErrorFrom(ErrAlreadyExists)
	case isErrOrPanic(err):
		return util.NewErrorFrom(err)
	}
	return nil
}

func (v *Secret) update(tx *sql.Tx) error {
	v.CreatedAt = time.Now().UTC()
	if err := v.validate(true); err != nil {
		return err
	}
	_, err := v.dbInsert(tx)
	switch {
	case IsDuplicateErr(err):
		return util.NewErrorFrom(ErrAlreadyExists)
	case isErrOrPanic(err):
		return util.NewErrorFrom(err)
	}
	return nil
}

func (v Secret) validate(fistInsert bool) error {
	errs := util.NewErrorFields().(*util.Error)
	if len(v.Id) == 0 {
		errs.SetFieldError("id", "missing")
	}
	if len(v.Team) == 0 {
		errs.SetFieldError("team", "missing")
	}
	if len(v.Id) < 10 {
		errs.SetFieldError("id", "invalid")
	}
	if len(v.Data) < 32 {
		errs.SetFieldError("data", "invalid")
	}
	if v.VaultVersion == 0 {
		errs.SetFieldError("vault_version", "invalid")
	}
	if v.Version == 0 {
		errs.SetFieldError("version", "invalid")
	}
	return errs.SetErrorOrCamo(ErrInvalidAttributes)
}
