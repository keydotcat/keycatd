package managers

import (
	"encoding/json"

	"github.com/keydotcat/keycatd/models"
)

type BroadcastAction string

const (
	BCAST_ACTION_SECRET_NEW    = BroadcastAction("secret:new")
	BCAST_ACTION_SECRET_CHANGE = BroadcastAction("secret:change")
	BCAST_ACTION_SECRET_REMOVE = BroadcastAction("secret:remove")
)

type Broadcast struct {
	Team    string
	Vault   string
	Message []byte
}

type BroadcasterMgr interface {
	Subscribe(address string) <-chan *Broadcast
	Unsubscribe(address string)
	Send(team, vault string, action BroadcastAction, secret *models.Secret)
	Stop()
}

type BroadcastPayload struct {
	Team   string          `json:"team"`
	Vault  string          `json:"vault"`
	Action BroadcastAction `json:"action"`
	Secret *models.Secret  `json:"secret,omitempty"`
}

func createBroadcast(team, vault string, action BroadcastAction, secret *models.Secret) *Broadcast {
	msg, err := json.Marshal(BroadcastPayload{team, vault, action, secret})
	if err != nil {
		panic(err)
	}
	return &Broadcast{team, vault, msg}
}
