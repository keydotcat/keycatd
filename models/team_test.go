package models

import "testing"

func TestCreateTeam(t *testing.T) {
	ctx := getCtx()
	owner := getDummyUser()
	vkp := VaultKeyPair{make([]byte, 32), map[string][]byte{owner.Id: []byte("crap")}}
	tName := owner.Id + " other team"
	team1, err := owner.CreateTeam(ctx, tName, vkp)
	if err != nil {
		t.Fatal(err)
	}
	team2, err := owner.CreateTeam(ctx, tName, vkp)
	if err != nil {
		t.Fatal(err)
	}
	if team1.Id == team2.Id {
		t.Errorf("Team IDs match! '%s' == '%s'", team1.Id, team2.Id)
	}
}
