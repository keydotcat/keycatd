package api

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/keydotcat/backend/models"
)

func TestGetAllSecrets(t *testing.T) {
	u := loginDummyUser()
	ctx := getCtx()
	teams, err := u.GetTeams(ctx)
	if err != nil {
		t.Fatal(err)
	}
	team := teams[0]
	r, err := GetRequest(fmt.Sprintf("/team/%s/secret", team.Id))
	CheckErrorAndResponse(t, r, err, 200)
	sga := &teamSecretGetAllResponse{}
	if err := json.NewDecoder(r.Body).Decode(sga); err != nil {
		t.Fatal(err)
	}
	if len(sga.Secrets) > 0 {
		t.Fatalf("Unexpected number of secrets: 0 vs %d", len(sga.Secrets))
	}
	vs := &vaultListResponse{}
	r, err = GetRequest(fmt.Sprintf("/team/%s/vault", team.Id))
	CheckErrorAndResponse(t, r, err, 200)
	if err := json.NewDecoder(r.Body).Decode(vs); err != nil {
		t.Fatal(err)
	}
	if len(vs.Vaults) != 1 {
		t.Fatalf("Unexpected number of vaults: 1 vs %d", len(vs.Vaults))
	}
	v := vs.Vaults[0]
	vPriv := unsealVaultKey(&v.Vault, v.Key)
	vcsr := &vaultCreateSecretRequest{Data: signAndPack(vPriv, a32b), Meta: signAndPack(vPriv, a32b)}
	r, err = PostRequest(fmt.Sprintf("/team/%s/vault/%s/secret", team.Id, v.Vault.Id), vcsr)
	CheckErrorAndResponse(t, r, err, 200)
	s := &models.Secret{}
	if err := json.NewDecoder(r.Body).Decode(s); err != nil {
		t.Fatal(err)
	}
	r, err = GetRequest(fmt.Sprintf("/team/%s/secret", team.Id))
	CheckErrorAndResponse(t, r, err, 200)
	sga = &teamSecretGetAllResponse{}
	if err := json.NewDecoder(r.Body).Decode(sga); err != nil {
		t.Fatal(err)
	}
	if len(sga.Secrets) > 1 {
		t.Fatalf("Unexpected number of secrets: 0 vs %d", len(sga.Secrets))
	}
}
