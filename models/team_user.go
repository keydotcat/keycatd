package models

import (
	"database/sql"

	"github.com/keydotcat/server/util"
)

type teamUser struct {
	Team           string `scaneo:"pk" json:"-"`
	User           string `scaneo:"pk" json:"user"`
	Admin          bool   `json:"admin"`
	AccessRequired bool   `json:"-"`
}

func (tu *teamUser) insert(tx *sql.Tx) error {
	_, err := tu.dbInsert(tx)
	if IsDuplicateErr(err) {
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
