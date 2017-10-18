package models

import "github.com/lib/pq"

func isDuplicateErr(err error) bool {
	pe, ok := err.(*pq.Error)
	return ok && pe.Code == "23505"
}
