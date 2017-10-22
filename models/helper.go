package models

import (
	"database/sql"

	"github.com/lib/pq"
)

func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	pe, ok := err.(*pq.Error)
	return ok && pe.Code == "23505"
}

func isNotExistsErr(err error) bool {
	return err == sql.ErrNoRows
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
