package models

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/keydotcat/backend/util"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
)

var a32b = make([]byte, 32)

func getDummyUser() *User {
	ctx := getCtx()
	uid := "u_" + util.GenerateRandomToken(10)
	ppub, ppriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	pub := (*ppub)[:]
	priv := (*ppriv)[:]
	vkp := getDummyVaultKeyPair(priv, uid)
	u, _, err := NewUser(ctx, uid, "uid fullname", uid+"@nowhere.net", uid, pub, signAndPack(priv, priv), vkp)
	if err != nil {
		panic(err)
	}
	return u
}

func signAndPack(key []byte, msg []byte) []byte {
	sig := ed25519.Sign(ed25519.PrivateKey(key), msg)
	response := make([]byte, SignatureSize+len(msg))
	copy(response[:SignatureSize], sig)
	copy(response[SignatureSize:], msg)
	return response
}

func getDummyVaultKeyPair(signer []byte, ids ...string) VaultKeyPair {
	ppub, ppriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	pub := (*ppub)[:]
	priv := (*ppriv)[:]
	vkp := VaultKeyPair{signAndPack(signer, pub), map[string][]byte{}}
	for _, id := range ids {
		vkp.Keys[id] = signAndPack(priv, priv)
	}
	return vkp
}

func TestCreateUser(t *testing.T) {
	ctx := getCtx()
	vkp := VaultKeyPair{make([]byte, 32), map[string][]byte{"test": []byte("crap")}}
	u, tok, err := NewUser(ctx, "test", "easdsa", "asdas@asdas.com", "somepass", make([]byte, 32), make([]byte, 32), vkp)
	if err != nil {
		fmt.Println(util.GetStack(err))
		t.Fatal(err)
	}
	if err = u.CheckPassword("somepass"); err != nil {
		fmt.Println("Password didn't check")
	}
	if u.Id != "test" {
		fmt.Println(util.GetStack(err))
		t.Errorf("Invalid username: %s vs test", u.Id)
	}
	u, tok, err = NewUser(ctx, "test", "easdsa", "asdas@asdas.com", "somepass", make([]byte, 32), make([]byte, 32), vkp)
	if err != nil && !util.CheckErr(err, ErrAlreadyExists) {
		t.Fatal(err)
	}
	if tok.Type != TOKEN_VERIFICATION || tok.User != u.Id {
		t.Fatalf("Token verification incomplete")
	}
	teams, err := u.GetTeams(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(teams) != 1 {
		t.Fatalf("Expected to have 1 team and got %d", len(teams))
	}
	team := teams[0]
	if team.Name != u.FullName {
		t.Errorf("Team name mismatch expected %s and got %s", u.FullName, team.Name)
	}
	if team.Owner != u.Id {
		t.Errorf("Team owner mismatch expected %s and got %s", u.Id, team.Owner)
	}
	vaults, err := team.GetVaultsForUser(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	if len(vaults) != 1 {
		t.Fatalf("Expected to have 1 vaults and got %d", len(vaults))
	}
	vault := vaults[0]
	if vault.Id != DEFAULT_VAULT_NAME {
		t.Errorf("Team name mismatch expected %s and got %s", DEFAULT_VAULT_NAME, vault.Id)
	}
	nu, err := FindUser(ctx, u.Id)
	if err != nil {
		t.Fatal(err)
	}
	if nu == nil {
		t.Fatalf("Could not find user by email")
	}
	if nu.Id != u.Id {
		t.Errorf("Mismatch in user IDs. Got %s and expected %s", nu.Id, u.Id)
	}
	nu, err = FindUserByEmail(ctx, u.Email)
	if err != nil {
		t.Fatal(err)
	}
	if nu == nil {
		t.Fatalf("Could not find user by email")
	}
	if nu.Id != u.Id {
		t.Errorf("Mismatch in user IDs. Got %s and expected %s", nu.Id, u.Id)
	}
}
