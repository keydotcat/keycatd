package api

import (
	"encoding/json"
	"testing"

	"github.com/keydotcat/keycatd/models"
	"github.com/keydotcat/keycatd/util"
)

func TestGetAllTeams(t *testing.T) {
	u := loginDummyUser()
	ctx := getCtx()
	teams, err := u.GetTeams(ctx)
	if err != nil {
		t.Fatal(err)
	}
	r, err := GetRequest("/team")
	CheckErrorAndResponse(t, r, err, 200)
	sga := &teamGetAllResponse{}
	if err := json.NewDecoder(r.Body).Decode(sga); err != nil {
		t.Fatal(err)
	}
	if len(teams) != len(sga.Teams) {
		t.Fatalf("Unexpected number of teams: %d vs %d", len(teams), len(sga.Teams))
	}
	privKeys := getUserPrivateKeys(u.PublicKey, u.Key)
	vkp := getDummyVaultKeyPair(privKeys, u.Id)
	tcr := teamCreateRequest{util.GenerateRandomToken(5), vkp}
	r, err = PostRequest("/team", tcr)
	CheckErrorAndResponse(t, r, err, 200)
	tf := &models.TeamFull{}
	if err := json.NewDecoder(r.Body).Decode(tf); err != nil {
		t.Fatal(err)
	}
	r, err = GetRequest("/team")
	CheckErrorAndResponse(t, r, err, 200)
	sga = &teamGetAllResponse{}
	if err := json.NewDecoder(r.Body).Decode(sga); err != nil {
		t.Fatal(err)
	}
	if len(teams)+1 != len(sga.Teams) {
		t.Fatalf("Unexpected number of teams: %d vs %d", len(teams)+1, len(sga.Teams))
	}
}
