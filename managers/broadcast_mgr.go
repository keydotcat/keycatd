package managers

import (
	"encoding/json"

	"github.com/keydotcat/server/models"
)

type Broadcast struct {
	Team    string
	Vault   string
	Message []byte
}

type BroadcasterMgr interface {
	Subscribe(address string) <-chan *Broadcast
	Unsubscribe(address string)
	Send(team, vault, action string, secret *models.Secret)
	Stop()
}

type BroadcastPayload struct {
	Team   string         `json:"team"`
	Vault  string         `json:"vault"`
	Action string         `json:"action"`
	Secret *models.Secret `json:"secret,omitempty"`
}

func createBroadcast(team, vault, action string, secret *models.Secret) *Broadcast {
	msg, err := json.Marshal(BroadcastPayload{team, vault, action, secret})
	if err != nil {
		panic(err)
	}
	return &Broadcast{team, vault, msg}
}
