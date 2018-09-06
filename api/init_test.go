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
	"github.com/keydotcat/backend/thelpers"
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
		DB:       thelpers.GetDBConnString(),
		DBType:   thelpers.GetTestDBType(),
		MailFrom: "blackhole@key.cat",
		SessionRedis: &ConfSessionRedis{
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
	thelpers.DropAllTables(apiH.db)
	m := db.NewMigrateMgr(apiH.db, thelpers.GetTestDBType())
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
