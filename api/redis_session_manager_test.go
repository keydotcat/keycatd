package api

import (
	"fmt"
	"testing"

	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

func init() {
	rs, err := NewRedisSessionManager("localhost:6379")
	if err != nil {
		panic(err)
	}
	if err = rs.(redisSessionManager).purgeAllData(); err != nil {
		panic(err)
	}
}

func addSessionForUser(rs SessionManager, u *models.User, agent string) error {
	s, err := rs.NewSession(u.Id, agent, false)
	if err != nil {
		return err
	}
	if s.UserId != u.Id {
		return fmt.Errorf("Mismatch in the user id: %s vs %s", s.UserId, u.Id)
	}
	return nil
}

func TestRedisSessionManager(t *testing.T) {
	rs, err := NewRedisSessionManager("localhost:6379")
	if err != nil {
		t.Fatal(err)
	}
	u := getDummyUser()
	u2 := getDummyUser()
	if err := addSessionForUser(rs, u, u.Id+":s1"); err != nil {
		t.Fatal(err)
	}
	if err := addSessionForUser(rs, u, u.Id+":s2"); err != nil {
		t.Fatal(err)
	}
	if err := addSessionForUser(rs, u2, u2.Id+":s1"); err != nil {
		t.Fatal(err)
	}
	sess, err := rs.GetAllSessions(u.Id)
	if err != nil {
		t.Fatal(err)
	}
	sids := map[string]bool{u.Id + ":s1": false, u.Id + ":s2": false}
	for _, ses := range sess {
		sids[ses.Agent] = true
	}
	for k, v := range sids {
		if !v {
			t.Errorf("Didn't find session %s", k)
		}
	}
	s, err := rs.UpdateSession(sess[0].Id, sess[0].Agent+":")
	if err != nil {
		t.Fatal(err)
	}
	if s.Id != sess[0].Id || s.Agent != sess[0].Agent+":" {
		t.Fatalf("Session hasn't been updated")
	}
	_, err = rs.UpdateSession("nonexistant", "asd")
	if !util.CheckErr(err, models.ErrDoesntExist) {
		t.Fatalf("Unexpected error: %s vs %s", models.ErrDoesntExist, err)
	}
	if err = rs.DeleteSession(sess[0].Id); err != nil {
		t.Fatal(err)
	}
	_, err = rs.UpdateSession(sess[0].Id, "asd")
	if !util.CheckErr(err, models.ErrDoesntExist) {
		t.Fatalf("Unexpected error: %s vs %s", models.ErrDoesntExist, err)
	}
}
