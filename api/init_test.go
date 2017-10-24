package api

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"net/http/httptest"

	"golang.org/x/crypto/bcrypt"

	"github.com/keydotcat/backend/db"
	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

var mdb *sql.DB
var a32b = make([]byte, 32)
var srv httptest.Server
var apiH apiHandler

func init() {
	initDB()
	initSRV()
}

func initSRV() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	rs, err := NewRedisSessionManager("localhost:6379")
	if err != nil {
		panic(err)
	}
	apiH = NewAPIHander(
		mdb,
		rs,
		NewCSRF([]byte("4d018d7e070ca9d5da7e767001bdaf90"), []byte("4e3797182c94f05b384c81ed0246f6b4")),
	).(apiHandler)
	srv = httptest.Server{
		Listener: ln,
		Config:   &http.Server{Handler: apiH},
	}
	srv.Start()
	log.Printf("Starting test server %s", srv.URL)
}

func initDB() {
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
