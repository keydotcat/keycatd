package managers

type MailMgr interface {
	SendMail(to string, subject string, data string) error
}
