package models

import (
	"context"
	"database/sql"
	"log"

	"github.com/keydotcat/backend/db"
	"github.com/keydotcat/backend/thelpers"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var mdb *sql.DB

var DUMP_BEFORE_TEST = false

func init() {
	HASH_PASSWD_COST = bcrypt.MinCost
	mdb = thelpers.GetDBConn()
	thelpers.DropAllTables(mdb)
	m := db.NewMigrateMgr(mdb, thelpers.GetTestDBType())
	if err := m.LoadMigrations(); err != nil {
		panic(err)
	}
	lid, ap, err := m.ApplyRequiredMigrations()
	if err != nil {
		panic(err)
	}
	log.Printf("Executed migrations until %d (%d applied)", lid, ap)
}

func getCtx() context.Context {
	return AddDBToContext(context.Background(), mdb)
}
