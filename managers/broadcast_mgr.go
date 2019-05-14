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
	BCAST_ACTION_VAULT_VERSION = BroadcastAction("vault:version")
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
	Action       BroadcastAction              `json:"action"`
	Team         string                       `json:"team,omitempty"`
	Vault        string                       `json:"vault,omitempty"`
	Secret       *models.Secret               `json:"secret,omitempty"`
	VaultVersion map[string]map[string]uint32 `json:"vault_version,omitempty"`
}

func createBroadcast(team, vault string, action BroadcastAction, secret *models.Secret) *Broadcast {
	msg, err := json.Marshal(BroadcastPayload{action, team, vault, secret, nil})
	if err != nil {
		panic(err)
	}
	return &Broadcast{team, vault, msg}
}
