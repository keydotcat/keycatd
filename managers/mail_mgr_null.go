package managers

import "io"

func NewMailMgrNULL() MailMgr {
	return mailMgrNULL(false)
}

type mailMgrNULL bool

func (s mailMgrNULL) SendMail(to, subject, data string) error {
	return nil
}

func (s mailMgrNULL) sendHeaders(to, subject string, sink io.WriteCloser) error {
	return nil
}
