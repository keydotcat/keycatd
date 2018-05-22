package models

import (
	"context"
	"database/sql"
	"log"

	"github.com/keydotcat/backend/db"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var mdb *sql.DB

var DUMP_BEFORE_TEST = false

func init() {
	HASH_PASSWD_COST = bcrypt.MinCost
	var err error
	mdb, err = sql.Open("postgres", "user=root dbname=test sslmode=disable port=26257")
	if err != nil {
		panic(err)
	}
	if DUMP_BEFORE_TEST {
		tables := []string{}
		rows, err := mdb.Query("SHOW TABLES")
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		var name string
		for rows.Next() {
			if err := rows.Scan(&name); err != nil {
				panic(err)
			}
			tables = append(tables, name)
		}
		for _, tname := range tables {
			log.Printf("Dropping %s", tname)
			if _, err = mdb.Exec("DROP TABLE \"" + tname + "\" CASCADE"); err != nil {
				panic(err)
			}
		}
	}
	m := db.NewMigrateMgr(mdb)
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
