package models

import (
	"testing"

	"github.com/keydotcat/backend/util"
)

func getFirstVault(o *User, t *Team) *Vault {
	vs, err := t.GetVaultsForUser(getCtx(), o)
	if err != nil {
		panic(err)
	}
	return vs[0]
}

func TestAddModifyAndDeleteSecret(t *testing.T) {
	ctx := getCtx()
	o, team := getDummyOwnerWithTeam()
	v := getFirstVault(o, team)
	s := &Secret{Data: a32b}
	if err := v.AddSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	if len(s.Id) < 10 {
		t.Errorf("Missing required secret id")
	}
	if s.Team != v.Team {
		t.Errorf("Mismatch in the team")
	}
	if s.Vault != v.Id {
		t.Errorf("Mismatch in the secret vault")
	}
	if err := v.UpdateSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	if s.CreatedAt.Equal(s.UpdatedAt) {
		t.Fatal("Didn't update the updated at time")
	}
	secrets, err := v.GetSecrets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ss := range secrets {
		if ss.Id == s.Id {
			found = true
			break
		}
	}
	if !found {
		t.Error("Could not find stored secret")
	}
	if err := v.DeleteSecret(ctx, s.Id); err != nil {
		t.Fatal(err)
	}
	if err := v.UpdateSecret(ctx, s); !util.CheckErr(err, ErrDoesntExist) {
		t.Fatalf("Expected different error: %s vs %s", ErrDoesntExist, err)
	}
}
