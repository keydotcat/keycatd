package managers

import (
	"fmt"
	"testing"

	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

func init() {
	rs, err := NewSessionMgrRedis("localhost:6379", 10)
	if err != nil {
		panic(err)
	}
	if err = rs.(sessionMgrRedis).purgeAllData(); err != nil {
		panic(err)
	}
}

func addSessionForUser(rs SessionMgr, uid, agent string) error {
	s, err := rs.NewSession(uid, agent, false)
	if err != nil {
		return err
	}
	if s.UserId != uid {
		return fmt.Errorf("Mismatch in the user id: %s vs %s", s.UserId, uid)
	}
	return nil
}

func TestRedisSessionManager(t *testing.T) {
	rs, err := NewSessionMgrRedis("localhost:6379", 10)
	if err != nil {
		t.Fatal(err)
	}
	uid1 := util.GenerateRandomToken(4)
	uid2 := util.GenerateRandomToken(4)
	if err := addSessionForUser(rs, uid1, uid1+":s1"); err != nil {
		t.Fatal(err)
	}
	if err := addSessionForUser(rs, uid1, uid1+":s2"); err != nil {
		t.Fatal(err)
	}
	if err := addSessionForUser(rs, uid2, uid2+":s1"); err != nil {
		t.Fatal(err)
	}
	sess, err := rs.GetAllSessions(uid1)
	if err != nil {
		t.Fatal(err)
	}
	sids := map[string]bool{uid1 + ":s1": false, uid1 + ":s2": false}
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
