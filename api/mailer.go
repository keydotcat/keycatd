package api

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sync"

	"github.com/keydotcat/backend/managers"
	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/static"
	"github.com/keydotcat/backend/util"
)

type mailer struct {
	templatesDir string
	rootUrl      string
	lock         *sync.Mutex
	TestMode     bool
	t            *template.Template
	mailMgr      managers.MailMgr
}

func newMailer(rootUrl string, testMode bool, mm managers.MailMgr) (*mailer, error) {
	m := &mailer{}
	m.templatesDir = "mail"
	m.rootUrl = rootUrl
	m.lock = &sync.Mutex{}
	m.mailMgr = mm
	if err := m.compile(); err != nil {
		return nil, err
	}
	return m, nil
}

func (mm *mailer) compile() error {
	mm.lock.Lock()
	defer mm.lock.Unlock()
	mm.t = template.New("mail_base")
	return static.Walk(mm.templatesDir, func(path string, info os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if ext == ".tmpl" {
			buf, err := static.Asset(path)
			if err != nil {
				return util.NewErrorFrom(err)
			}
			_, err = mm.t.New(filepath.Base(path[0 : len(path)-len(ext)])).Parse(string(buf))
			if err != nil {
				return util.NewErrorFrom(err)
			}
		}
		return nil
	})
}

type mailUserTeamTokenData struct {
	FullName string
	HostUrl  string
	Team     string
	Token    string
	Email    string
	Username string
}

func (mm *mailer) send(muttd mailUserTeamTokenData, templateName, subject string) error {
	if mm.TestMode {
		return nil
	}
	buf := util.BufPool.Get()
	defer util.BufPool.Put(buf)
	err := mm.t.ExecuteTemplate(buf, templateName, muttd)
	if err != nil {
		panic(err)
	}
	return mm.mailMgr.SendMail(muttd.Email, subject, buf.String())
}

func (mm *mailer) sendConfirmationMail(u *models.User, token *models.Token) error {
	email := u.Email
	if u.UnconfirmedEmail != "" {
		email = u.UnconfirmedEmail
	}
	muttd := mailUserTeamTokenData{FullName: u.FullName, HostUrl: mm.rootUrl, Token: token.Id, Username: u.Id, Email: email}
	return mm.send(muttd, "confirm_account", "Confirm your email")
}

func (mm *mailer) sendInvitationMail(t *models.Team, u *models.User, i *models.Invite) error {
	muttd := mailUserTeamTokenData{FullName: u.FullName, HostUrl: mm.rootUrl, Email: i.Email, Team: t.Name}
	return mm.send(muttd, "invite_user", fmt.Sprintf("%s has invited you to join key.cat", u.FullName))
}
