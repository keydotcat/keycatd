package models

import (
	"testing"

	"github.com/keydotcat/backend/util"
)

func TestConfirmEmail(t *testing.T) {
	ctx := getCtx()
	uid := "u_" + util.GenerateRandomToken(10)
	_, priv, pack := generateNewKeys()
	vkp := getDummyVaultKeyPair(priv, uid)
	u, tok, err := NewUser(ctx, uid, "uid fullname", uid+"@nowhere.net", uid, pack, vkp)
	if err != nil {
		panic(err)
	}
	u2, err := tok.ConfirmEmail(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if u2.Id != u.Id {
		t.Errorf("Mismatch in user id")
	}
	if u2.UnconfirmedEmail != "" {
		t.Errorf("Unconfirmed email didn't reset")
	}
	if u2.Email != uid+"@nowhere.net" {
		t.Errorf("Email mismatch")
	}
	if !u2.ConfirmedAt.Valid {
		t.Fatalf("User is not confirmed")
	}
	_, err = FindToken(ctx, tok.Id)
	if !util.CheckErr(err, ErrDoesntExist) {
		t.Fatalf("Unexpected error: %s vs %s", ErrDoesntExist, err)
	}
	u3, err := FindUser(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if !u3.ConfirmedAt.Valid {
		t.Fatalf("User is not confirmed in the db")
	}

}
