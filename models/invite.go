package models

import (
	"database/sql"
	"time"

	"github.com/keydotcat/backend/util"
)

type Invite struct {
	Team      string `scaneo:"pk"`
	Email     string `scaneo:"pk"`
	CreatedAt time.Time
}

func (i Invite) validate() error {
	errs := util.NewErrorFields().(*util.Error)
	if !reValidEmail.MatchString(i.Email) {
		errs.SetFieldError("email", "invalid")
	}
	if len(i.Team) == 0 {
		errs.SetFieldError("team", "invalid")
	}
	return errs.Camo()
}

func (u *Invite) insert(tx *sql.Tx) error {
	if err := u.validate(); err != nil {
		return err
	}
	u.CreatedAt = time.Now().UTC()
	_, err := u.dbInsert(tx)
	if isDuplicateErr(err) {
		return util.NewErrorFrom(ErrAlreadyInvited)
	}
	if isErrOrPanic(err) {
		return err
	}
	return nil
}
