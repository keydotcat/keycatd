package models

import (
	"testing"

	"github.com/keydotcat/backend/util"
)

func getDummyOwnerWithTeam() (*User, *Team) {
	ctx := getCtx()
	owner := getDummyUser()
	ts, err := owner.GetTeams(ctx)
	if err != nil {
		panic(err)
	}
	if len(ts) != 1 {
		panic("Got more than 1 team")
	}
	return owner, ts[0]
}

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

func TestInviteUserToTeam(t *testing.T) {
	ctx := getCtx()
	owner, team := getDummyOwnerWithTeam()
	added, err := team.AddOrInviteUserByEmail(ctx, owner, "a@a.com")
	if err != nil {
		t.Fatal(err)
	}
	if added {
		t.Fatalf("Added user when it had to be invited")
	}
	added, err = team.AddOrInviteUserByEmail(ctx, owner, "a@a.com")
	if !util.CheckErr(err, ErrAlreadyInvited) {
		t.Fatalf("Expected error %s and got %s", ErrAlreadyInvited, err)
	}
}

func TestAddExistingUserToTeam(t *testing.T) {
	ctx := getCtx()
	owner, team := getDummyOwnerWithTeam()
	invitee := getDummyUser()
	added, err := team.AddOrInviteUserByEmail(ctx, owner, invitee.Email)
	if err != nil {
		t.Fatal(err)
	}
	if !added {
		t.Fatalf("Added user when it had to be invited")
	}
	added, err = team.AddOrInviteUserByEmail(ctx, owner, invitee.Email)
	if !util.CheckErr(err, ErrAlreadyInTeam) {
		t.Fatalf("Expected error %s and got %s", ErrAlreadyInTeam, err)
	}
}
