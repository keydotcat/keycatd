package api

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sync"

	"github.com/keydotcat/server/managers"
	"github.com/keydotcat/server/models"
	"github.com/keydotcat/server/static"
	"github.com/keydotcat/server/util"
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
			_, err = mm.t.New(path[len(mm.templatesDir)+1 : len(path)-len(ext)]).Parse(string(buf))
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

func (mm *mailer) send(muttd mailUserTeamTokenData, locale, templateName, subject string) error {
	if mm.TestMode {
		return nil
	}
	buf := util.BufPool.Get()
	defer util.BufPool.Put(buf)
	tpl := mm.t.Lookup(fmt.Sprintf("%s/%s", locale, templateName))
	if tpl == nil {
		tpl = mm.t.Lookup("en/" + templateName)
	}
	if tpl == nil {
		panic("No template found with name " + templateName)
	}
	err := tpl.Execute(buf, muttd)
	if err != nil {
		panic(err)
	}
	return mm.mailMgr.SendMail(muttd.Email, subject, buf.String())
}

func (mm *mailer) sendConfirmationMail(u *models.User, token *models.Token, locale string) error {
	email := u.Email
	if u.UnconfirmedEmail != "" {
		email = u.UnconfirmedEmail
	}
	muttd := mailUserTeamTokenData{FullName: u.FullName, HostUrl: mm.rootUrl, Token: token.Id, Username: u.Id, Email: email}
	return mm.send(muttd, locale, "confirm_account", "Confirm your email")
}

func (mm *mailer) sendInvitationMail(t *models.Team, u *models.User, i *models.Invite, locale string) error {
	muttd := mailUserTeamTokenData{FullName: u.FullName, HostUrl: mm.rootUrl, Email: i.Email, Team: t.Name}
	return mm.send(muttd, locale, "invite_user", fmt.Sprintf("%s has invited you to join key.cat", u.FullName))
}

func (mm *mailer) sendTestEmail(to string) error {
	muttd := mailUserTeamTokenData{Email: to}
	return mm.send(muttd, "en", "test_email", "Keycat test email")
}

func SendTestEmail(c Conf, to string) error {
	err := c.validate()
	if err != nil {
		return err
	}
	var m *mailer
	switch {
	case c.MailSMTP != nil:
		m, err = newMailer(c.Url, TEST_MODE, managers.NewMailMgrSMTP(c.MailSMTP.Server, c.MailSMTP.User, c.MailSMTP.Password, c.MailFrom))
	case c.MailSparkpost != nil:
		m, err = newMailer(c.Url, TEST_MODE, managers.NewMailMgrSparkpost(c.MailSparkpost.Key, c.MailFrom, c.MailSparkpost.EU))
	default:
		return util.NewErrorf("No mail was configured")
	}
	if err != nil {
		return util.NewErrorf("Could not create mailer: %s", err)
	}
	return m.sendTestEmail(to)
}
