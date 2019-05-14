package models

import (
	"testing"

	"github.com/keydotcat/keycatd/util"
)

func XgetFirstVault(o *User, t *Team) *VaultFull {
	vs, err := t.GetVaultsFullForUser(getCtx(), o)
	if err != nil {
		panic(err)
	}
	return vs[0]
}

func TestAddModifyAndDeleteSecret(t *testing.T) {
	ctx := getCtx()
	o, team := getDummyOwnerWithTeam()
	vm := getFirstVault(o, team)
	s := &Secret{Data: signAndPack(vm.priv, a32b)}
	version := vm.v.Version
	if err := vm.v.AddSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	if vm.v.Version != version+1 {
		t.Fatal("Vault version didn't increase")
	}
	if len(s.Id) < 10 {
		t.Errorf("Missing required secret id")
	}
	if s.Team != vm.v.Team {
		t.Errorf("Mismatch in the team")
	}
	if s.Vault != vm.v.Id {
		t.Errorf("Mismatch in the secret vault")
	}
	if s.VaultVersion != vm.v.Version {
		t.Fatalf("Mismatch in the vault (%d) and secret vault (%d) version", vm.v.Version, s.VaultVersion)
	}
	if s.Version != 1 {
		t.Fatalf("Invalid secret version, expected 1 and got %d", s.Version)
	}
	if err := vm.v.UpdateSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	if s.Version != 2 {
		t.Fatalf("Invalid secret version, expected 2 and got %d", s.Version)
	}
	if vm.v.Version != version+2 {
		t.Fatal("Vault version didn't increase")
	}
	if s.VaultVersion != vm.v.Version {
		t.Fatalf("Mismatch in the vault (%d) and secret vault (%d) version", vm.v.Version, s.VaultVersion)
	}
	secrets, err := vm.v.GetSecrets(ctx)
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
	if err := vm.v.DeleteSecret(ctx, s.Id); err != nil {
		t.Fatal(err)
	}
	if err := vm.v.UpdateSecret(ctx, s); !util.CheckErr(err, ErrDoesntExist) {
		t.Fatalf("Expected different error: %s vs %s", ErrDoesntExist, err)
	}
}

func TestAddSecretList(t *testing.T) {
	ctx := getCtx()
	o, team := getDummyOwnerWithTeam()
	vm := getFirstVault(o, team)
	sl := []*Secret{
		&Secret{Data: signAndPack(vm.priv, a32b)},
		&Secret{Data: signAndPack(vm.priv, a32b)},
	}
	if err := vm.v.AddSecretList(ctx, sl); err != nil {
		t.Fatal(err)
	}
	for i, s := range sl {
		if len(s.Id) == 0 {
			t.Errorf("Secret %d does not have an id", i)
		}
	}
	if sl[0].VaultVersion != vm.v.Version-1 {
		t.Fatalf("Mismatch in the vault (%d) and secret vault (%d) version", vm.v.Version-1, sl[0].VaultVersion)
	}
	if sl[1].VaultVersion != vm.v.Version {
		t.Fatalf("Mismatch in the vault (%d) and secret vault (%d) version", vm.v.Version, sl[1].VaultVersion)
	}
}
