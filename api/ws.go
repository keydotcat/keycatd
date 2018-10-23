package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
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

func getTeamVaultMapForUser(ctx context.Context, u *models.User) (map[string][]string, error) {
	teams, err := u.GetTeams(ctx)
	if err != nil {
		return nil, err
	}
	tv := map[string][]string{}
	for _, t := range teams {
		vs, err := t.GetVaultsForUser(ctx, u)
		if err != nil {
			return nil, err
		}
		tv[t.Id] = make([]string, len(vs))
		for i, v := range vs {
			tv[t.Id][i] = v.Id
		}
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

// /ws
func (ah apiHandler) wsSubscribe(w http.ResponseWriter, r *http.Request) error {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return util.NewErrorFrom(err)
	}
	defer ws.Close()
	go receiveWsPongs(ws)
	ctx := r.Context()
	currentUser := ctxGetUser(ctx)
	tv, err := getTeamVaultMapForUser(ctx, currentUser)
	if err != nil {
		return err
	}
	bChan := ah.bcast.Subscribe(r.RemoteAddr)
	defer ah.bcast.Unsubscribe(r.RemoteAddr)
	alive := true
	for alive {
		select {
		case <-time.After(2 * time.Minute):
			if err := ws.WriteMessage(websocket.PingMessage, []byte{1}); err != nil {
				alive = false
			}
		case b := <-bChan:
			vs, ok := tv[b.Team]
			if !ok {
				continue
			}
			found := false
			for _, v := range vs {
				if v == b.Vault {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			if err := ws.WriteMessage(websocket.TextMessage, b.Message); err != nil {
				alive = false
			}
		}
	}
	return nil
}
