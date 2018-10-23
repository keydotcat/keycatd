package api

import (
	"encoding/json"
	"testing"

	"github.com/keydotcat/keycatd/models"
	"github.com/keydotcat/keycatd/util"
)

func loginDummyUser() *models.User {
	u := getDummyUser()
	s, err := apiH.sm.NewSession(u.Id, "1.1.1.1", "none", true)
	if err != nil {
		panic(err)
	}
	activeSessionToken = s.Id
	return u
}

func TestRegister(t *testing.T) {
	activeSessionToken = ""
	uid := util.GenerateRandomToken(5)
	_, priv, fullpack := generateNewKeys()
	vkp := getDummyVaultKeyPair(priv, uid)
	arp := authRegisterRequest{
		uid,
		util.GenerateRandomToken(10) + "@me.not",
		"Random name",
		"pass",
		fullpack,
		vkp.PublicKey,
		vkp.Keys[uid],
	}
	r, err := PostRequest("/auth/register", arp)
	CheckErrorAndResponse(t, r, err, 200)
	ar := authRequest{Id: arp.Username, Password: arp.Password, RequireCSRF: true}
	r, err = PostRequest("/auth/login", ar)
	CheckErrorAndResponse(t, r, err, 401)
	tokens := models.FindTokensForUser(getCtx(), arp.Username)
	if len(tokens) != 1 {
		t.Fatalf("Expected to find one token and found %d", len(tokens))
	}
	r, err = GetRequest("/auth/confirm_email/" + tokens[0].Id)
	CheckErrorAndResponse(t, r, err, 200)
	u := &models.User{}
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		t.Fatal(err)
	}
	if u.Id != arp.Username {
		t.Fatalf("Mismatch in the user id!: %s vs %s", arp.Username, u.Id)
	}
}

func TestLogin(t *testing.T) {
	activeSessionToken = ""
	u := getDummyUser()
	ar := authRequest{Id: u.Id, Password: u.Id, RequireCSRF: true}
	r, err := PostRequest("/auth/login", ar)
	CheckErrorAndResponse(t, r, err, 200)
	s := &authLoginResponse{}
	if err := json.NewDecoder(r.Body).Decode(s); err != nil {
		t.Fatal(err)
	}
	if s.Username != u.Id {
		t.Fatalf("Mismatch in the user id!: %s vs %s", u.Id, s.Username)
	}
	activeCsrfToken = s.Csrf
}
