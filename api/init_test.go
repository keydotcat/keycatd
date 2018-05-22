package api

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/codahale/http-handlers/logging"
	"github.com/keydotcat/backend/db"
	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

var srv httptest.Server
var apiH apiHandler

func init() {
	TEST_MODE = true
	initSRV()
	initDB()
}

func initSRV() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	c := Conf{
		Port:     1, //Not used
		Url:      "http://" + ln.Addr().String(),
		DB:       "user=root dbname=test sslmode=disable port=26257",
		MailFrom: "blackhole@key.cat",
		SessionRedis: ConfSessionRedis{
			Server: "localhost:6379",
			DBId:   10,
		},
		Csrf: ConfCsrf{
			HashKey:  "4d018d7e070ca9d5da7e767001bdaf90",
			BlockKey: "4e3797182c94f05b384c81ed0246f6b4",
		},
	}
	handler, err := NewAPIHandler(c)
	if err != nil {
		panic(err)
	}
	apiH = handler.(apiHandler)
	logHandler := logging.Wrap(apiH, os.Stdout)
	logHandler.Start()
	srv = httptest.Server{
		Listener: ln,
		Config:   &http.Server{Handler: logHandler},
	}
	srv.Start()
	log.Printf("Starting test server %s", srv.URL)
}

var DUMP_BEFORE_TEST = false

func initDB() {
	models.HASH_PASSWD_COST = bcrypt.MinCost
	var err error
	if DUMP_BEFORE_TEST {
		tables := []string{}
		rows, err := apiH.db.Query("SHOW TABLES")
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
			if _, err = apiH.db.Exec("DROP TABLE \"" + tname + "\" CASCADE"); err != nil {
				panic(err)
			}
		}
	}
	m := db.NewMigrateMgr(apiH.db)
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
	return models.AddDBToContext(context.Background(), apiH.db)
}

func getDummyUser() *models.User {
	ctx := getCtx()
	uid := util.GenerateRandomToken(5)
	_, priv, fullpack := generateNewKeys()
	vkp := getDummyVaultKeyPair(priv, uid)
	u, t, err := models.NewUser(ctx, uid, "uid fullname", uid+"@nowhere.net", uid, fullpack, vkp)
	if err != nil {
		panic(err)
	}
	if _, err = t.ConfirmEmail(getCtx()); err != nil {
		panic(err)
	}
	return u
}
