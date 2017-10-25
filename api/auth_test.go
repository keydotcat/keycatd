package api

import (
	"encoding/json"
	"testing"

	"github.com/keydotcat/backend/managers"
	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

func loginDummyUser() {
	u := getDummyUser()
	s, err := apiH.sm.NewSession(u.Id, "none", true)
	if err != nil {
		panic(err)
	}
	activeSessionToken = s.Id
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
	tok := &models.Token{}
	if err := json.NewDecoder(r.Body).Decode(tok); err != nil {
		t.Fatal(err)
	}
	r, err = GetRequest("/auth/confirm_email/" + tok.Id)
	CheckErrorAndResponse(t, r, err, 200)
	u := &models.User{}
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		t.Fatal(err)
	}
	if u.Id != arp.Id {
		t.Fatalf("Mismatch in the user id!: %s vs %s", arp.Id, u.Id)
	}
}

func TestLogin(t *testing.T) {
	u := getDummyUser()
	ar := authRequest{Id: u.Id, Password: u.Id, RequireCSRF: true}
	r, err := PostRequest("/auth/login", ar)
	CheckErrorAndResponse(t, r, err, 200)
	s := &managers.Session{}
	if err := json.NewDecoder(r.Body).Decode(s); err != nil {
		t.Fatal(err)
	}
	if s.UserId != u.Id {
		t.Fatalf("Mismatch in the user id!: %s vs %s", u.Id, s.UserId)
	}
}
