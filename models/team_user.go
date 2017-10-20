package models

import (
	"database/sql"

	"github.com/keydotcat/backend/util"
)

type teamUser struct {
	Team           string `scaneo:"pk"`
	User           string `scaneo:"pk"`
	Admin          bool
	AccessRequired bool
}

func (tu *teamUser) insert(tx *sql.Tx) error {
	_, err := tu.dbInsert(tx)

	if err != nil {
		if isDuplicateErr(err) {
			return util.NewErrorf("User %s is already in team", tu.User)
		}
		return util.NewErrorf("Could not add user to team: %s", err)
	}
	return nil
}

func (tu *teamUser) update(tx *sql.Tx) error {
	_, err := tu.dbUpdate(tx)

	if err != nil {
		if isDuplicateErr(err) {
			return util.NewErrorf("User %s is already in team", tu.User)
		}
		return util.NewErrorf("Could not update user in team: %s", err)
	}
	return nil
}
