package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetAndDeleteSessions(t *testing.T) {
	u := loginDummyUser()
	s, err := apiH.sm.NewSession(u.Id, "none", true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GETSES")
	r, err := GetRequest("/session/" + activeSessionToken)
	CheckErrorAndResponse(t, r, err, 200)
	sr := &sessionGetTokenResponse{}
	if err := json.NewDecoder(r.Body).Decode(sr); err != nil {
		t.Fatal(err)
	}
	if len(sr.StoreToken) == 0 {
		t.Errorf("Expected to get store tokens")
	}
	r, err = GetRequest("/session/" + s.Id)
	CheckErrorAndResponse(t, r, err, 200)
	sr = &sessionGetTokenResponse{}
	if err := json.NewDecoder(r.Body).Decode(sr); err != nil {
		t.Fatal(err)
	}
	if len(sr.StoreToken) > 0 {
		t.Errorf("Expected NOT to get the store token %#v", sr)
	}
	r, err = DeleteRequest("/session/" + s.Id)
	CheckErrorAndResponse(t, r, err, 200)
	r, err = GetRequest("/session/" + s.Id)
	CheckErrorAndResponse(t, r, err, 404)
}
