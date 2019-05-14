package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/keydotcat/keycatd/managers"
	"github.com/keydotcat/keycatd/models"
	"github.com/keydotcat/keycatd/util"
)

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ah apiHandler) wsRoot(w http.ResponseWriter, r *http.Request) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		return ah.wsSubscribe(w, r)
	}
	return util.NewErrorFrom(ErrNotFound)
}

func getTeamVaultMapForUser(ctx context.Context, u *models.User) (map[string][]*models.Vault, error) {
	teams, err := u.GetTeams(ctx)
	if err != nil {
		return nil, err
	}
	tv := map[string][]*models.Vault{}
	for _, t := range teams {
		vs, err := t.GetVaultsForUser(ctx, u)
		if err != nil {
			return nil, err
		}
		tv[t.Id] = vs
	}
	return tv, nil
}

func receiveWsPongs(ws *websocket.Conn) {
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

type eventSender interface {
	sendMessage([]byte) error
	sendPing() error
}

func (ah apiHandler) broadcastEventListenLoop(r *http.Request, eb eventSender) error {
	ctx := r.Context()
	currentUser := ctxGetUser(ctx)
	tv, err := getTeamVaultMapForUser(ctx, currentUser)
	if err != nil {
		return err
	}
	buf := util.BufPool.Get()
	verMsg := managers.BroadcastPayload{
		Action:       managers.BCAST_ACTION_VAULT_VERSION,
		VaultVersion: map[string]map[string]uint32{},
	}
	for tid, vaults := range tv {
		verMsg.VaultVersion[tid] = map[string]uint32{}
		for _, vault := range vaults {
			verMsg.VaultVersion[tid][vault.Id] = vault.Version
		}
	}
	if err := json.NewEncoder(buf).Encode(verMsg); err != nil {
		return err
	}
	defer util.BufPool.Put(buf)
	if err := eb.sendMessage(buf.Bytes()); err != nil {
		return err
	}
	bChan := ah.bcast.Subscribe(r.RemoteAddr)
	defer ah.bcast.Unsubscribe(r.RemoteAddr)
	alive := true
	for alive {
		select {
		case <-time.After(time.Second * 30):
			if err := eb.sendPing(); err != nil {
				alive = false
			}
		case b := <-bChan:
			vs, ok := tv[b.Team]
			if !ok {
				continue
			}
			found := false
			for _, v := range vs {
				if v.Id == b.Vault {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			if err := eb.sendMessage(b.Message); err != nil {
				alive = false
			}
		}
	}
	return nil
}

type webSocketSender struct {
	ws *websocket.Conn
}

func (ss webSocketSender) sendPing() error {
	return ss.ws.WriteMessage(websocket.PingMessage, []byte{1})
}

func (ss webSocketSender) sendMessage(msg []byte) error {
	return ss.ws.WriteMessage(websocket.TextMessage, msg)
}

// /ws
func (ah apiHandler) wsSubscribe(w http.ResponseWriter, r *http.Request) error {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return util.NewErrorFrom(err)
	}
	defer ws.Close()
	go receiveWsPongs(ws)
	return ah.broadcastEventListenLoop(r, webSocketSender{ws})
}
