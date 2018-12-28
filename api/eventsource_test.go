package api

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/keydotcat/keycatd/managers"
)

func TestGetEventSourceNotifications(t *testing.T) {
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
	resp, err := EventRequest("/eventsource")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		vPriv := unsealVaultKey(&v.Vault, v.Key)
		vcsr := &vaultCreateSecretRequest{Data: signAndPack(vPriv, a32b)}
		r, err := PostRequest(fmt.Sprintf("/team/%s/vault/%s/secret", teams[0].Id, v.Vault.Id), vcsr)
		CheckErrorAndResponse(t, r, err, 200)
	}()
	bp := &managers.BroadcastPayload{}
	if err := json.NewDecoder(resp.Body).Decode(bp); err != nil {
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
