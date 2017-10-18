package models

import "github.com/lib/pq"

func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	pe, ok := err.(*pq.Error)
	return ok && pe.Code == "23505"
}
