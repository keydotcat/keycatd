package api

import (
	"encoding/json"
	"testing"
)

func TestGetAndDeleteSessions(t *testing.T) {
	u := loginDummyUser()
	s, err := apiH.sm.NewSession(u.Id, "none", true)
	if err != nil {
		t.Fatal(err)
	}
	r, err := GetRequest("/session/" + activeSessionToken)
	CheckErrorAndResponse(t, r, err, 200)
	sr := &sessionResponse{}
	if err := json.NewDecoder(r.Body).Decode(sr); err != nil {
		t.Fatal(err)
	}
	if len(sr.Csrf) == 0 || len(sr.StoreToken) == 0 {
		t.Errorf("Expected to get the Csrf and store tokens")
	}
	r, err = GetRequest("/session/" + s.Id)
	CheckErrorAndResponse(t, r, err, 200)
	sr = &sessionResponse{}
	if err := json.NewDecoder(r.Body).Decode(sr); err != nil {
		t.Fatal(err)
	}
	if len(sr.Csrf) > 0 || len(sr.StoreToken) > 0 {
		t.Errorf("Expected NOT to get the Csrf and store tokens %s", sr)
	}
	r, err = DeleteRequest("/session/" + s.Id)
	CheckErrorAndResponse(t, r, err, 200)
	r, err = GetRequest("/session/" + s.Id)
	CheckErrorAndResponse(t, r, err, 404)
}
