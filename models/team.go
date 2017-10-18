package models

import (
	"database/sql"
	"time"

	"github.com/keydotcat/backend/util"
)

type Team struct {
	Id            string
	Name          string
	Owner         string
	BelongsToUser bool
	Size          int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func createUserTeam(tx *sql.Tx, owner *User, name string) (*Team, error) {
	now := time.Now().UTC()
	t := &Team{
		util.GenerateRandomToken(16),
		name,
		owner.Id,
		true,
		0,
		now,
		now,
	}
	if err := t.insert(tx); err != nil {
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
