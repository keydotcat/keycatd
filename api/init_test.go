package api

import (
	"context"
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"

	"github.com/keydotcat/backend/db"
	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

var mdb *sql.DB
var a32b = make([]byte, 32)

func init() {
	models.HASH_PASSWD_COST = bcrypt.MinCost
	var err error
	mdb, err = sql.Open("postgres", "user=root dbname=test sslmode=disable port=26257")
	if err != nil {
		panic(err)
	}
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
	m := db.NewMigrateMgr(mdb)
	if err := m.LoadMigrations(); err != nil {
		panic(err)
	}
	lid, err := m.ApplyRequiredMigrations()
	if err != nil {
		panic(err)
	}
	log.Printf("Executed migrations until %d", lid)
}

func getCtx() context.Context {
	return models.AddDBToContext(context.Background(), mdb)
}

func getDummyUser() *models.User {
	ctx := getCtx()
	uid := "u_" + util.GenerateRandomToken(10)
	vkp := getDummyVaultKeyPair(uid)
	u, _, err := models.NewUser(ctx, uid, "uid fullname", uid+"@nowhere.net", uid, a32b, a32b, vkp)
	if err != nil {
		panic(err)
	}
	return u
}

func getDummyVaultKeyPair(ids ...string) models.VaultKeyPair {
	vkp := models.VaultKeyPair{a32b, map[string][]byte{}}
	for _, id := range ids {
		vkp.Keys[id] = a32b
	}
	return vkp
}
