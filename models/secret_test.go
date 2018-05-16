package models

import (
	"testing"

	"github.com/keydotcat/backend/util"
)

func TestGetAllSecretsForUser(t *testing.T) {
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
	s2 := &Secret{Data: signAndPack(v2Priv, a32b), Meta: signAndPack(v2Priv, a32b)}
	err = v2.AddSecret(ctx, s2)
	if err != nil {
		t.Fatal(err)
	}
	vPriv := unsealVaultKey(&vf.Vault, vf.Key)
	s := &Secret{Data: signAndPack(vPriv, a32b), Meta: signAndPack(vPriv, a32b)}
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
