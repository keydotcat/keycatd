package models

import (
	"testing"
)

func TestGetAllSecretsForOwnerAndUser(t *testing.T) {
	ctx := getCtx()
	owner, team := getDummyOwnerWithTeam()
	vm := getFirstVault(owner, team)
	vm2 := createVaultMock(owner, team)
	invitee := getDummyUser()
	_, err := team.AddOrInviteUserByEmail(ctx, owner, invitee.Email)
	if err != nil {
		t.Fatal(err)
	}
	uk := map[string][]byte{invitee.Id: sealVaultKey(vm2.v, vm2.priv)}
	err = vm2.v.AddUsers(ctx, uk)
	if err != nil {
		t.Fatal(err)
	}
	s2 := &Secret{Data: signAndPack(vm2.priv, a32b)}
	err = vm2.v.AddSecret(ctx, s2)
	if err != nil {
		t.Fatal(err)
	}
	s := &Secret{Data: signAndPack(vm.priv, a32b)}
	err = vm.v.AddSecret(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	secsInv, err := team.GetSecretsForUser(ctx, invitee)
	if err != nil {
		t.Fatal(err)
	}
	if len(secsInv) != 1 {
		t.Errorf("Expected 1 secret and got %d", len(secsInv))
	}
	secsOwner, err := team.GetSecretsForUser(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(secsOwner) != 2 {
		t.Errorf("Expected 2 secret and got %d", len(secsOwner))
	}
}

func TestUpdateSecret(t *testing.T) {
	ctx := getCtx()
	owner, team := getDummyOwnerWithTeam()
	vm := getFirstVault(owner, team)
	s := &Secret{Data: signAndPack(vm.priv, a32b)}
	if err := vm.v.AddSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	s.Version = 0
	s.VaultVersion = 0
	if err := vm.v.UpdateSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	if s.Version != 2 {
		t.Errorf("Unexpected secret version 2 vs %d", s.Version)
	}
	if s.VaultVersion != vm.v.Version {
		t.Errorf("Mismatch in the vault version %d vs %d", s.VaultVersion, vm.v.Version)
	}
	if err := vm.v.UpdateSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	if s.Version != 3 {
		t.Errorf("Unexpected secret version 2 vs %d", s.Version)
	}
}

func TestCheckRetrieveLastVersionOfSecret(t *testing.T) {
	ctx := getCtx()
	owner, team := getDummyOwnerWithTeam()
	vm := getFirstVault(owner, team)
	vm2 := createVaultMock(owner, team)
	secrets := map[string]*Secret{}
	var vault *Vault
	numSecrets := 10
	var s *Secret
	for i := 0; i < numSecrets; i++ {
		if i%2 == 0 {
			vault = vm.v
			s = &Secret{Data: signAndPack(vm.priv, a32b)}
		} else {
			vault = vm2.v
			s = &Secret{Data: signAndPack(vm2.priv, a32b)}
		}
		if err := vault.AddSecret(ctx, s); err != nil {
			t.Fatal(err)
		}
		for j := 0; j < i; j++ {
			if err := vault.UpdateSecret(ctx, s); err != nil {
				t.Fatal(err)
			}
		}
		secrets[s.Id] = s
	}
	secs, err := team.GetSecretsForUser(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(secs) != numSecrets {
		t.Fatalf("Expected %d secrets and got %d", numSecrets, len(secs))
	}
	for _, s := range secs {
		rs, ok := secrets[s.Id]
		switch {
		case !ok:
			t.Fatal("Unknown secret")
		case s.Version != rs.Version:
			t.Errorf("Expected version %d and got %d", rs.Version, s.Version)
		}
	}
}

func TestMoveSecretToVault(t *testing.T) {
	ctx := getCtx()
	owner, team := getDummyOwnerWithTeam()
	vm := getFirstVault(owner, team)
	vms := []vaultMock{vm}
	for len(vms) < 3 {
		vms = append(vms, createVaultMock(owner, team))
	}
	secrets := make([]*Secret, 3)
	for i, vm := range vms[:2] {
		s := &Secret{Data: signAndPack(vm.priv, a32b)}
		secrets[i] = s
		if err := vm.v.AddSecret(ctx, s); err != nil {
			t.Fatal(err)
		}
		if err := vm.v.UpdateSecret(ctx, s); err != nil {
			t.Fatal(err)
		}
	}
	if err := secrets[0].MoveToTeamVault(ctx, team.Id, vms[2].v.Id); err != nil {
		t.Fatal(err)
	}
	secrets[2], secrets[0] = secrets[0], secrets[2]
	for i, vm := range vms {
		vs, err := vm.v.GetSecretsAllVersions(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if secrets[i] == nil {
			if len(vs) != 0 {
				t.Fatalf("Expected 0 secrets for vault %d and got %d", i, len(vs))
			}
		} else {
			if len(vs) != 2 {
				t.Fatalf("Expected 2 secrets for vault %d and got %d", i, len(vs))
			}
			for si := range vs {
				if secrets[i].Id != vs[si].Id {
					t.Fatalf("Secret id mismatch for vault %d", i)
				}
			}
		}
	}
}

func TestMoveSecretToTeamVault(t *testing.T) {
	ctx := getCtx()
	owner, team := getDummyOwnerWithTeam()
	teams := []*Team{team, createTeamMock(owner)}
	vms := make([][]vaultMock, len(teams))
	secrets := make([][]*Secret, len(teams))
	for it, team := range teams {
		secrets[it] = make([]*Secret, 3)
		vms[it] = []vaultMock{getFirstVault(owner, team)}
		for len(vms[it]) < 3 {
			vms[it] = append(vms[it], createVaultMock(owner, team))
		}
		for iv, vm := range vms[it][:2] {
			s := &Secret{Data: signAndPack(vm.priv, a32b)}
			secrets[it][iv] = s
			if err := vm.v.AddSecret(ctx, s); err != nil {
				t.Fatal(err)
			}
			if err := vm.v.UpdateSecret(ctx, s); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := secrets[0][0].MoveToTeamVault(ctx, teams[1].Id, vms[1][2].v.Id); err != nil {
		t.Fatal(err)
	}
	secrets[1][2], secrets[0][0] = secrets[0][0], secrets[1][2]
	for it := range vms {
		for iv, vm := range vms[it] {
			vs, err := vm.v.GetSecretsAllVersions(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if secrets[it][iv] == nil {
				if len(vs) != 0 {
					t.Fatalf("Expected 0 secrets for vault %d/%d and got %d", it, iv, len(vs))
				}
			} else {
				if len(vs) != 2 {
					t.Fatalf("Expected 2 secrets for vault %d/%d and got %d", it, iv, len(vs))
				}
				for si := range vs {
					if secrets[it][iv].Id != vs[si].Id {
						t.Fatalf("Secret id mismatch for vault %d/%d", it, iv)
					}
				}
			}
		}
	}
}
