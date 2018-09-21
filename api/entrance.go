package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/keydotcat/server/db"
	"github.com/keydotcat/server/managers"
	"github.com/keydotcat/server/models"
	"github.com/keydotcat/server/util"
)

var TEST_MODE = false

type apiOptions struct {
	onlyInvited bool
}

type apiHandler struct {
	db            *sql.DB
	sm            managers.SessionMgr
	mail          *mailer
	csrf          csrf
	staticHandler *StaticHandler
	options       apiOptions
}

func NewAPIHandler(c Conf) (http.Handler, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}
	ah := apiHandler{}
	ah.options.onlyInvited = c.OnlyInvited
	ah.db, err = sql.Open("postgres", c.DB)
	if err != nil {
		return nil, util.NewErrorf("Could not connect to db '%s': %s", c.DB, err)
	}
	ah.db.SetMaxOpenConns(c.DBMaxConns)
	m := db.NewMigrateMgr(ah.db, c.DBType)
	if err := m.LoadMigrations(); err != nil {
		panic(err)
	}
	lid, ap, err := m.ApplyRequiredMigrations()
	if err != nil {
		fmt.Println(util.GetStack(err))
		panic(err)
	}
	log.Printf("Executed migrations until %d (%d applied)", lid, ap)
	switch {
	case TEST_MODE:
		ah.mail, err = newMailer(c.Url, TEST_MODE, managers.NewMailMgrNULL())
	case c.MailSMTP != nil:
		ah.mail, err = newMailer(c.Url, TEST_MODE, managers.NewMailMgrSMTP(c.MailSMTP.Server, c.MailSMTP.User, c.MailSMTP.Password, c.MailFrom))
	case c.MailSparkpost != nil:
		ah.mail, err = newMailer(c.Url, TEST_MODE, managers.NewMailMgrSparkpost(c.MailSparkpost.Key, c.MailFrom, c.MailSparkpost.EU))
	default:
	}
	if err != nil {
		return nil, util.NewErrorf("Could not create mailer: %s", err)
	}
	if c.SessionRedis != nil {
		ah.sm, err = managers.NewSessionMgrRedis(c.SessionRedis.Server, c.SessionRedis.DBId)
		if err != nil {
			return nil, util.NewErrorf("Could not connect to redis at %s: %s", c.SessionRedis.Server, err)
		}
	} else {
		ah.sm = managers.NewSessionMgrDB(ah.db)
	}
	var blockKey []byte
	if len(c.Csrf.BlockKey) > 0 {
		blockKey = []byte(c.Csrf.BlockKey)
	}
	ah.csrf = newCsrf([]byte(c.Csrf.HashKey), blockKey)
	ah.staticHandler = NewStaticHandler()
	return ah, nil
}

func (ah apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(models.AddDBToContext(r.Context(), ah.db))
	head, subPath := shiftPath(r.URL.Path)
	if head == "api" {
		r.URL.Path = subPath
		ah.apiRoot(w, r)
	} else {
		ah.staticHandler.ServeHTTP(w, r)
	}
}

func (ah apiHandler) apiRoot(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(models.AddDBToContext(r.Context(), ah.db))
	head := ""
	head, r.URL.Path = shiftPath(r.URL.Path)
	if head == "auth" {
		if err := ah.authRoot(w, r); err != nil {
			httpErr(w, err)
		}
		return
	}
	err := util.NewErrorFrom(ErrNotFound)
	r = ah.authorizeRequest(w, r)
	if r == nil {
		return
	}
	//From here on you need to be authenticated
	switch head {
	case "session":
		err = ah.sessionRoot(w, r)
	case "user":
		err = ah.userRoot(w, r)
	case "team":
		err = ah.teamRoot(w, r)
	}
	if err != nil {
		httpErr(w, err)
	}
}
