package managers

import (
	"io"
	"net/smtp"
	"strings"

	"github.com/keydotcat/keycatd/util"
)

func NewMailMgrSMTP(server, user, pass, from string) MailMgr {
	return mailMgrSMTP{server, user, pass, from}
}

type mailMgrSMTP struct {
	Server   string
	User     string
	Password string
	From     string
}

func (s mailMgrSMTP) SendMail(to, subject, data string) error {
	c, err := smtp.Dial(s.Server)
	if err != nil {
		return err
	}
	defer c.Quit()

	if len(s.User) > 0 {
		host := strings.Split(s.Server, ":")[0]
		if err = c.Auth(smtp.PlainAuth("", s.User, s.Password, host)); err != nil {
			panic(err)
		}
	}

	// Set the sender and recipient.
	if err = c.Mail(s.From); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()
	if err = s.sendHeaders(to, subject, wc); err != nil {
		return err
	}
	if _, err = io.WriteString(wc, data); err != nil {
		return err
	}
	return nil
}

func (s mailMgrSMTP) sendHeaders(to, subject string, sink io.WriteCloser) error {
	buf := util.BufPool.Get()
	defer util.BufPool.Put(buf)
	buf.Write([]byte("From: " + s.From + "\r\n"))
	buf.Write([]byte("To: " + to + "\r\n"))
	buf.Write([]byte("Subject: " + subject + "\r\n"))
	buf.Write([]byte("MIME-Version: 1.0\r\n"))
	buf.Write([]byte("Content-Type: text/html; charset=\"utf-8\"\r\n"))
	buf.Write([]byte("\r\n"))
	_, err := buf.WriteTo(sink)
	return err
}
