package models

import (
	"database/sql"
	"time"

	"github.com/keydotcat/server/util"
)

type Invite struct {
	Team      string    `scaneo:"pk" json:"-"`
	Email     string    `scaneo:"pk" json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func findInvitesForUser(tx *sql.Tx, email string) ([]*Invite, error) {
	rows, err := tx.Query(`SELECT `+selectInviteFields+` FROM "invite" WHERE "email" = $1`, email)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	invites, err := scanInvites(rows)
	isErrOrPanic(err)
	return invites, util.NewErrorFrom(err)

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
	if IsDuplicateErr(err) {
		return util.NewErrorFrom(ErrAlreadyInvited)
	}
	if isErrOrPanic(err) {
		return err
	}
	return nil
}

func (i *Invite) getTeam(tx *sql.Tx) (*Team, error) {
	t := &Team{}
	r := tx.QueryRow(`SELECT `+selectTeamFullFields+` FROM "team" WHERE "team"."id" = $1`, i.Team)
	err := t.dbScanRow(r)
	if isNotExistsErr(err) {
		return nil, util.NewErrorFrom(ErrDoesntExist)
	}
	isErrOrPanic(err)
	return t, err
}
