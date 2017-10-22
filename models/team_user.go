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
	if isDuplicateErr(err) {
		return util.NewErrorFrom(ErrAlreadyInTeam)
	}
	if isErrOrPanic(err) {
		return util.NewErrorFrom(err)
	}
	return nil
}

func (tu *teamUser) update(tx *sql.Tx) error {
	_, err := tu.dbUpdate(tx)
	if isErrOrPanic(err) {
		return util.NewErrorFrom(err)
	}
	return nil
}
