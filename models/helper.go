package models

import (
	"database/sql"
	"regexp"

	"github.com/keydotcat/keycatd/util"
	"github.com/lib/pq"
)

var regUniqueFieldMessage = regexp.MustCompile(`\Aduplicate key value \((\w+)\).*\z`)
var regUniqueFieldDetailFull = regexp.MustCompile(`\AKey \((\w+)\)=\(\w+\) already exists.*\z`)
var regUniqueFieldDetail = regexp.MustCompile(`\((\w+)\)=\(\w+\)`)

func IsDuplicateErr(err error) bool {
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
	if pe.Code == "23505" {
		if f := regUniqueFieldDetail.FindStringSubmatch(pe.Detail); len(f) > 1 {
			return f[1]
		}
		return pe.Constraint
	}
	//Fallback
	if f := regUniqueFieldMessage.FindStringSubmatch(pe.Message); len(f) > 1 {
		return f[1]
	}
	if f := regUniqueFieldDetail.FindStringSubmatch(pe.Detail); len(f) > 1 {
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
