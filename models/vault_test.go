package models

import (
	"testing"

	"github.com/keydotcat/backend/util"
)

func getFirstVault(o *User, t *Team) *VaultFull {
	vs, err := t.GetVaultsFullForUser(getCtx(), o)
	if err != nil {
		panic(err)
	}
	return vs[0]
}

func TestAddModifyAndDeleteSecret(t *testing.T) {
	ctx := getCtx()
	o, team := getDummyOwnerWithTeam()
	v := getFirstVault(o, team)
	vPriv := unsealVaultKey(&v.Vault, v.Key)
	s := &Secret{Data: signAndPack(vPriv, a32b)}
	version := v.Version
	if err := v.AddSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	if v.Version != version+1 {
		t.Fatal("Vault version didn't increase")
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
	if s.VaultVersion != v.Version {
		t.Fatalf("Mismatch in the vault (%d) and secret vault (%d) version", v.Version, s.VaultVersion)
	}
	if err := v.UpdateSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	if v.Version != version+2 {
		t.Fatal("Vault version didn't increase")
	}
	if s.CreatedAt.Equal(s.UpdatedAt) {
		t.Fatal("Didn't update the updated at time")
	}
	if s.VaultVersion != v.Version {
		t.Fatalf("Mismatch in the vault (%d) and secret vault (%d) version", v.Version, s.VaultVersion)
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
