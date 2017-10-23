package models

import (
	"database/sql"
	"time"

	"github.com/keydotcat/backend/util"
)

const TOKEN_VERIFICATION = "verification"

type Token struct {
	Id        string    `scaneo:"pk" json:"id"`
	Type      string    `json:"-"`
	User      string    `json:"-"`
	Extra     string    `json:"extra,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *Token) validate() error {
	errs := util.NewErrorFields().(*util.Error)
	if !reValidUsername.MatchString(u.User) {
		errs.SetFieldError("username", "invalid")
	}
	if len(u.Id) < 6 {
		errs.SetFieldError("id", "too short")
	}
	if u.Type != TOKEN_VERIFICATION {
		errs.SetFieldError("type", "invalid")
	}
	return errs.Camo()
}

func (u *Token) insert(tx *sql.Tx) error {
	u.Id = util.GenerateRandomToken(32)
	if err := u.validate(); err != nil {
		return err
	}
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = u.CreatedAt
	_, err := u.dbInsert(tx)
	if isDuplicateErr(err) {
		return util.NewErrorFrom(ErrAlreadyExists)
	}
	isErrOrPanic(err)
	return util.NewErrorFrom(err)
}

func (u *Token) update(tx *sql.Tx) error {
	if err := u.validate(); err != nil {
		return err
	}
	u.UpdatedAt = time.Now().UTC()
	res, err := u.dbUpdate(tx)
	return treatUpdateErr(res, err)
}
