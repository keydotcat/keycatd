package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
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
	source := bufio.NewReader(resp.Body)
	line, err := source.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(line, "id: ") != 0 {
		t.Fatalf("Didn't find 'id' in line %s", line)
	}
	line, err = source.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(line, "event: ") != 0 {
		t.Fatalf("Didn't find 'event' in line %s", line)
	}
	line, err = source.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(line, "data: ") != 0 {
		t.Fatalf("Didn't find 'data' in line %s", line)
	}
	fmt.Println(line[6:])
	if err := json.NewDecoder(bytes.NewBuffer([]byte(line[6:]))).Decode(bp); err != nil {
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
