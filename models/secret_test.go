package models

import (
	"testing"

	"github.com/keydotcat/keycatd/util"
)

func TestGetAllSecretsForOwnerAndUser(t *testing.T) {
	ctx := getCtx()
	owner, team := getDummyOwnerWithTeam()
	ownerPrivKeys := getUserPrivateKeys(owner.PublicKey, owner.Key)
	vf := getFirstVault(owner, team)
	vname := util.GenerateRandomToken(5)
	vkp := getDummyVaultKeyPair(ownerPrivKeys, owner.Id)
	v2, err := team.CreateVault(ctx, owner, vname, vkp)
	if err != nil {
		t.Fatal(err)
	}
	invitee := getDummyUser()
	_, err = team.AddOrInviteUserByEmail(ctx, owner, invitee.Email)
	if err != nil {
		t.Fatal(err)
	}
	v2Priv := unsealVaultKey(v2, vkp.Keys[owner.Id])
	uk := map[string][]byte{invitee.Id: vkp.Keys[owner.Id]}
	err = v2.AddUsers(ctx, uk)
	if err != nil {
		t.Fatal(err)
	}
	s2 := &Secret{Data: signAndPack(v2Priv, a32b)}
	err = v2.AddSecret(ctx, s2)
	if err != nil {
		t.Fatal(err)
	}
	vPriv := unsealVaultKey(&vf.Vault, vf.Key)
	s := &Secret{Data: signAndPack(vPriv, a32b)}
	err = vf.Vault.AddSecret(ctx, s)
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
	vf := getFirstVault(owner, team)
	v := &vf.Vault
	vPriv := unsealVaultKey(&vf.Vault, vf.Key)
	s := &Secret{Data: signAndPack(vPriv, a32b)}
	if err := v.AddSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	s.Version = 0
	s.VaultVersion = 0
	if err := v.UpdateSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	if s.Version != 2 {
		t.Errorf("Unexpected secret version 2 vs %d", s.Version)
	}
	if s.VaultVersion != v.Version {
		t.Errorf("Mismatch in the vault version %d vs %d", s.VaultVersion, v.Version)
	}
	if err := v.UpdateSecret(ctx, s); err != nil {
		t.Fatal(err)
	}
	if s.Version != 3 {
		t.Errorf("Unexpected secret version 2 vs %d", s.Version)
	}
}

func TestCheckRetrieveLastVersionOfSecret(t *testing.T) {
	ctx := getCtx()
	owner, team := getDummyOwnerWithTeam()
	vf := getFirstVault(owner, team)
	v := &vf.Vault
	vPriv := unsealVaultKey(&vf.Vault, vf.Key)
	ownerPrivKeys := getUserPrivateKeys(owner.PublicKey, owner.Key)
	vkp := getDummyVaultKeyPair(ownerPrivKeys, owner.Id)
	v2, err := team.CreateVault(ctx, owner, util.GenerateRandomToken(5), vkp)
	if err != nil {
		t.Fatal(err)
	}
	v2Priv := unsealVaultKey(v2, vkp.Keys[owner.Id])
	secrets := map[string]*Secret{}
	var vault *Vault
	numSecrets := 10
	var s *Secret
	for i := 0; i < numSecrets; i++ {
		if i%2 == 0 {
			vault = v
			s = &Secret{Data: signAndPack(vPriv, a32b)}
		} else {
			vault = v2
			s = &Secret{Data: signAndPack(v2Priv, a32b)}
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
