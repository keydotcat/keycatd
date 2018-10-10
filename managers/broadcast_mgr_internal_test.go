package managers

import (
	"testing"
)

func TestInternalBroadcasterMgr(t *testing.T) {
	testBroadcastMgr("internal", NewInternalBroadcasterMgr(), t)
}
