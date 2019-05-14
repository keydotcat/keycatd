package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/keydotcat/keycatd/managers"
)

func connectWs(path string, t *testing.T) *websocket.Conn {
	d := websocket.Dialer{}
	d.Jar = getCookieJar()
	headers := http.Header{}
	headers.Add("X-Csrf-Token", activeCsrfToken)
	if len(activeSessionToken) != 0 {
		headers.Add("Authorization", fmt.Sprintf("Bearer %s", activeSessionToken))
	}
	conn, _, err := d.Dial(strings.Replace(srv.URL+path, "http", "ws", 1), headers)
	if err != nil {
		t.Fatalf("Could not connect:%s", err)
	}
	return conn
}

func TestGetWSNotifications(t *testing.T) {
	u := loginDummyUser()
	ctx := getCtx()
	teams, err := u.GetTeams(ctx)
	if err != nil {
		t.Fatal(err)
	}
	vs, err := teams[0].GetVaultsFullForUser(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	v := vs[0]
	ws := connectWs("/ws", t)
	defer ws.Close()
	go func() {
		vPriv := unsealVaultKey(&v.Vault, v.Key)
		vcsr := &vaultCreateSecretRequest{Data: signAndPack(vPriv, a32b)}
		r, err := PostRequest(fmt.Sprintf("/team/%s/vault/%s/secret", teams[0].Id, v.Vault.Id), vcsr)
		CheckErrorAndResponse(t, r, err, 200)
	}()
	bp := &managers.BroadcastPayload{}
	if err := ws.ReadJSON(bp); err != nil {
		t.Fatalf("Could not read the msg: %s", err)
	}
	if bp.Action != managers.BCAST_ACTION_VAULT_VERSION {
		t.Errorf("Unexpected action: %s vs %s", managers.BCAST_ACTION_VAULT_VERSION, bp.Action)
	}
	if err := ws.ReadJSON(bp); err != nil {
		t.Fatalf("Could not read the msg: %s", err)
	}
	if bp.Team != teams[0].Id || bp.Vault != v.Id {
		t.Errorf("Mismatch either in the team or in vault: %s:%s vs %s:%s", teams[0].Id, v.Id, bp.Team, bp.Vault)
	}
	if bp.Action != managers.BCAST_ACTION_SECRET_NEW {
		t.Errorf("Unexpected action: %s vs %s", managers.BCAST_ACTION_SECRET_NEW, bp.Action)
	}
	if bp.Secret == nil {
		t.Errorf("Missing secret")
	}
}
