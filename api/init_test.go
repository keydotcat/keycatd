package api

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"

	"golang.org/x/crypto/bcrypt"

	"github.com/keydotcat/backend/db"
	"github.com/keydotcat/backend/managers"
	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

var srv httptest.Server
var apiH apiHandler

func init() {
	TEST_MODE = false
	initSRV()
	initDB()
}

func initSRV() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	rs, err := managers.NewSessionMgrRedis("localhost:6379")
	if err != nil {
		panic(err)
	}
	c := Conf{
		URL:      "http://localhost:" + ln.Port,
		DB:       "user=root dbname=test sslmode=disable port=26257",
		MailFrom: "blackhole@key.cat",
		MailSMTP: &ConfMailSMTP{
			Server: "localhost:1025",
		},
		SessionRedis: ConfSessionRedis{
			Server: "localhost:6379",
			DbId:   10,
		},
		Csrf: ConfCsrf{
			HashKey: "4d018d7e070ca9d5da7e767001bdaf90",
			BlobKey: "4e3797182c94f05b384c81ed0246f6b4",
		},
	}
	handler, err = NewAPIHander(c)
	if err != nil {
		panic(err)
	}
	apiH = handler.(apiHandler)
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
	u, _, err := models.NewUser(ctx, uid, "uid fullname", uid+"@nowhere.net", uid, fullpack, vkp)
	if err != nil {
		panic(err)
	}
	return u
}
