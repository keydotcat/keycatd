package models

import (
	"database/sql"
	"regexp"

	"github.com/keydotcat/backend/util"
	"github.com/lib/pq"
)

var regUniqueField = regexp.MustCompile(`\Aduplicate key value \((\w+)\).*\z`)

func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	pe, ok := err.(*pq.Error)
	return ok && pe.Code == "23505"
}

func getDuplicateFieldFromErr(err error) string {
	pe, ok := err.(*pq.Error)
	if !ok {
		panic("Error is not a pq.Error")
	}
	f := regUniqueField.FindStringSubmatch(pe.Message)
	if len(f) > 1 {
		return f[1]
	}
	return ""
}

func isNotExistsErr(err error) bool {
	return util.CheckErr(err, sql.ErrNoRows)
}

func isErrOrPanic(err error) bool {
	if err != nil {
		if err != sql.ErrTxDone || err != sql.ErrConnDone {
			panic("Could not execute sql statement: " + err.Error())
		}
		return true
	}
	return false
}

func treatUpdateErr(res sql.Result, err error) error {
	if isErrOrPanic(err) {
		return util.NewErrorFrom(err)
	}
	if n, err := res.RowsAffected(); n == 0 {
		return util.NewErrorFrom(ErrDoesntExist)
	} else if err != nil {
		return util.NewErrorFrom(err)
	}
	return nil
}
