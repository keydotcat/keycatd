package models

import (
	"testing"

	"github.com/keydotcat/backend/util"
)

func TestGetAllSecretsForUser(t *testing.T) {
	ctx := getCtx()
	owner, team := getDummyOwnerWithTeam()
	ownerPrivKeys := getUserPrivateKeys(owner.PublicKey, owner.Key)
	vkp := getDummyVaultKeyPair(ownerPrivKeys, owner.Id)
	firstVault := getFirstVault(owner, team)
	vname := util.GenerateRandomToken(5)
	_, err := team.CreateVault(ctx, owner, vname, vkp)
	if err != nil {
		t.Fatal(err)
	}
	invitee := getDummyUser()
	_, err = team.AddOrInviteUserByEmail(ctx, owner, invitee.Email)
	if err != nil {
		t.Fatal(err)
	}
	uk := map[string][]byte{invitee.Id: signAndPack(vkp.Keys[owner.Id], vkp.Keys[owner.Id])}
	err = firstVault.AddUsers(ctx, uk)
	if err != nil {
		t.Fatal(err)
	}

}
