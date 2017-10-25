package api

import (
	"fmt"
	"html/template"
	"models"
	"os"
	"path/filepath"
	"static"
	"sync"
	"util"

	"github.com/keydotcat/backend/managers"
)

type mailer struct {
	templatesDir string
	rootUrl      string
	lock         *sync.Mutex
	TestMode     bool
	t            *template.Template
	mailer       managers.MailMgr
}

func newMailer(rootUrl string, testMode bool, mm managers.MailMgr) (*mailer, error) {
	mm := &Mailer{}
	mm.templatesDir = "mail"
	mm.rootUrl = conf.RootURL
	mm.lock = &sync.Mutex{}
	if err := mm.compile(); err != nil {
		return nil, err
	}
	return mm, nil
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
	Fullname string
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

func (mm *mailer) sendConfirmationMail(u *models.User) error {
	email := u.Email
	if u.UncheckedEmail != "" {
		email = u.UncheckedEmail
	}
	muttd := mailUserTeamTokenData{Fullname: u.Fullname, HostUrl: mm.rootUrl, Token: u.ConfirmToken, Username: u.Username, Email: email}
	return mm.send(muttd, "confirm_account", "Confirm your email")
}

func (mm *mailer) sendForgottenPasswordMail(u *models.User) error {
	muttd := mailUserTeamTokenData{Fullname: u.Fullname, Username: u.Username, HostUrl: mm.rootUrl, Token: u.ForgottenToken, Email: u.Email}
	return mm.send(muttd, "forgotten_password", "Reset your password")
}

func (mm *mailer) sendInvitationMail(t *models.Team, u *models.User, i *models.Invite) error {
	muttd := mailUserTeamTokenData{Fullname: u.Fullname, HostUrl: mm.rootUrl, Token: i.Token, Email: i.Email, Team: t.Name}
	return mm.send(muttd, "invite_user", fmt.Sprintf("%s has invited you to join key.cat", u.Fullname))
}
