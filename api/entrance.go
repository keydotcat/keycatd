package api

import (
	"database/sql"
	"log"
	"net/http"
	"util"

	"github.com/keydotcat/backend/managers"
	"github.com/keydotcat/backend/models"
)

var TEST_MODE = false

type apiHandler struct {
	db   *sql.DB
	sm   managers.SessionMgr
	mail *mailer
	csrf CSRF
}

func NewAPIHander(c Conf) (http.Handler, error) {
	var err error
	ah := apiHandler{}
	ah.db, err = sql.Open("postgres", c.DB)
	if err != nil {
		return nil, util.NewErrorf("Could not connect to db '%s': %s", c.DB, err)
	}
	switch {
	case c.MailSMTP != nil:
		ah.mail = newMailer(c.URL, TEST_MODE, managers.NewMailMgrSMTP(c.MailSMTP.Server, c.MailSMTP.User, c.MailSMTP.Pass, c.MailFrom))
	case c.MailSparkpost != nil:
		ah.mail = newMailer(c.URL, TEST_MODE, managers.NewMailMgrSparkpost(c.MailSparkpost.Key, c.MailFrom))
	default:
		return nil, util.NewErrorf("No mail manager defined in the configuration")
	}
	ah.sm, err = manager.NewSessionMgrRedis(c.SessionRedis.Server, c.SessionRedis.DBId)
	if err != nil {
		return nil, util.NewErrorf("Could not connect to redis at %s: %s", c.SessionRedis.Server, err)
	}
	ah.csrf = newCsrf([]byte(c.Csrf.HashKey), []byte(c.Csrf.BlobKey))
	return ah, nil
}

func (ah apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, "-", r.RequestURI)
	r = r.WithContext(models.AddDBToContext(r.Context(), ah.db))
	head := ""
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "auth":
		ah.authRoot(w, r)
	default:
		http.NotFound(w, r)
	}
}
