package managers

import "testing"

func init() {
	rs, err := NewSessionMgrRedis("localhost:6379", 10)
	if err != nil {
		panic(err)
	}
	if err = rs.(sessionMgrRedis).purgeAllData(); err != nil {
		panic(err)
	}
}

func TestRedisSessionManager(t *testing.T) {
	rs, err := NewSessionMgrRedis("localhost:6379", 10)
	if err != nil {
		t.Fatal(err)
	}
	testSessionManager(rs, t, "redis")
}
