package managers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/keydotcat/keycatd/util"
)

func NewMailMgrSparkpost(key, from string, eu bool) MailMgr {
	return mailMgrSparkPost{key, from, eu}
}

type mailMgrSparkPost struct {
	Key  string
	From string
	EU   bool
}

type spAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type spRecipient struct {
	ReturnPath string    `json:"return_path,omitempty"`
	Address    spAddress `json:"address"`
}

type spContent struct {
	From    spAddress `json:"from"`
	Subject string    `json:"subject"`
	ReplyTo string    `json:"reply_to,omitempty"`
	Text    string    `json:"text,omitempty"`
	Html    string    `json:"html,omitempty"`
}

type spMail struct {
	Recipients []spRecipient `json:"recipients"`
	Content    spContent     `json:"content"`
}

func (s mailMgrSparkPost) SendMail(to, subject, data string) error {
	sm := spMail{
		Recipients: []spRecipient{spRecipient{Address: spAddress{Email: to, Name: to}}},
		Content: spContent{
			From:    spAddress{Email: s.From, Name: "Key.cat"},
			Subject: subject,
			Html:    data,
		},
	}
	reqBody := util.BufPool.Get()
	defer util.BufPool.Put(reqBody)
	if err := json.NewEncoder(reqBody).Encode(sm); err != nil {
		return err
	}
	var endpoint string
	if s.EU {
		endpoint = "https://api.eu.sparkpost.com/api/v1/transmissions"
	} else {
		endpoint = "https://api.sparkpost.com/api/v1/transmissions"
	}
	req, _ := http.NewRequest("POST", endpoint, reqBody)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", s.Key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return util.NewErrorFrom(err)
	}

	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return util.NewError(string(resp_body))
	}
	return nil
}
