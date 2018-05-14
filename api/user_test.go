package api

import (
	"encoding/json"
	"testing"

	"github.com/keydotcat/backend/models"
)

func TestGetUserInfo(t *testing.T) {
	u := loginDummyUser()
	r, err := GetRequest("/user")
	CheckErrorAndResponse(t, r, err, 200)
	uf := &models.UserFull{}
	if err := json.NewDecoder(r.Body).Decode(uf); err != nil {
		t.Fatal(err)
	}
	if uf.Id != u.Id {
		t.Errorf("Mismatch in the user. Expected %s and got %s", uf.Id, u.Id)
	}
}
