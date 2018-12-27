package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/keydotcat/keycatd/db"
	"github.com/keydotcat/keycatd/managers"
	"github.com/keydotcat/keycatd/models"
	"github.com/keydotcat/keycatd/util"
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
	bcast         managers.BroadcasterMgr
}

func NewAPIHandler(c Conf) (http.Handler, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}
	ah := apiHandler{}
	ah.bcast = managers.NewInternalBroadcasterMgr()
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
	var err error
	head := ""
	head, r.URL.Path = shiftPath(r.URL.Path)
	//This is the non authenticated root
	switch head {
	case "auth":
		err = ah.authRoot(w, r)
	case "version":
		err = ah.versionRoot(w, r)
	default:
		err = ah.authenticatedRoot(w, r, head)
	}
	if err != nil {
		httpErr(w, err)
	}
}

func (ah apiHandler) authenticatedRoot(w http.ResponseWriter, r *http.Request, head string) error {
	//From here on you need to be authenticated
	err := util.NewErrorFrom(ErrNotFound)
	r = ah.authorizeRequest(w, r)
	if r == nil {
		return nil
	}
	switch head {
	case "session":
		err = ah.sessionRoot(w, r)
	case "user":
		err = ah.userRoot(w, r)
	case "team":
		err = ah.teamRoot(w, r)
	case "ws":
		err = ah.wsRoot(w, r)
	case "eventsource":
		err = ah.eventSourceRoot(w, r)
	}
	return err
}
