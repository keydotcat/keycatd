package models

import (
	"database/sql"
	"time"

	"github.com/keydotcat/backend/util"
)

type Secret struct {
	Team      string    `scaneo:"pk" json:"-"`
	Vault     string    `scaneo:"pk" json:"-"`
	Id        string    `scaneo:"pk" json:"id"`
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (v *Secret) insert(tx *sql.Tx) error {
	v.Id = util.GenerateRandomToken(10)
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

func (u *Secret) update(tx *sql.Tx) error {
	if err := u.validate(); err != nil {
		return err
	}
	u.UpdatedAt = time.Now().UTC()
	res, err := u.dbUpdate(tx)
	return treatUpdateErr(res, err)
}

func (v Secret) validate() error {
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
	return errs.Camo()
}
