package managers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/keydotcat/keycatd/models"
)

func validateBcastMsg(b *Broadcast, t *testing.T) error {
	p := &BroadcastPayload{}
	if err := json.Unmarshal(b.Message, p); err != nil {
		return fmt.Errorf("Could not unmarsall: %s", err)
	}
	ti := 0
	n, err := fmt.Sscanf(b.Team, "team:%d", &ti)
	if n != 1 || err != nil {
		return fmt.Errorf("Could not parse team value %s: %s", b.Team, err)
	}
	if b.Team != p.Team {
		return fmt.Errorf("Team mismatch: %s vs %s", b.Team, p.Team)
	}
	vi := 0
	n, err = fmt.Sscanf(b.Vault, "vault:%d", &vi)
	if n != 1 || err != nil {
		return fmt.Errorf("Could not parse team value %s: %s", b.Team, err)
	}
	if b.Vault != p.Vault {
		return fmt.Errorf("Vault mismatch: %s vs %s", b.Vault, p.Vault)
	}
	if ti%2 == 0 {
		if p.Action != BCAST_ACTION_SECRET_NEW {
			return fmt.Errorf("Action mismatch: %s vs %s", BCAST_ACTION_SECRET_NEW, p.Action)
		}
		if p.Secret == nil {
			return fmt.Errorf("Secret is nil")
		}
	} else {
		if p.Action != BCAST_ACTION_SECRET_REMOVE {
			return fmt.Errorf("Action mismatch: %s vs %s", BCAST_ACTION_SECRET_REMOVE, p.Action)
		}
		if p.Secret != nil {
			return fmt.Errorf("Secret is not nil")
		}
	}
	return nil
}

func testBroadcastMgr(name string, bc BroadcasterMgr, t *testing.T) {
	clients := map[string]<-chan *Broadcast{}
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		id := strconv.Itoa(i)
		clients[id] = bc.Subscribe(id)
		go func(id string, num int, c <-chan *Broadcast) {
			for b := range c {
				err := validateBcastMsg(b, t)
				if err != nil {
					t.Errorf("Could not validate broadcast: %s", err)
				}
				num = num - 1
				if num == 0 {
					break
				}
			}
			bc.Unsubscribe(id)
			wg.Done()
		}(id, i+10, clients[id])
	}
	for i := 0; i < 1000; i++ {
		sec := &models.Secret{}
		if i%2 == 0 {
			bc.Send(fmt.Sprintf("team:%d", i), fmt.Sprintf("vault:%d", i), BCAST_ACTION_SECRET_NEW, sec)
		} else {
			bc.Send(fmt.Sprintf("team:%d", i), fmt.Sprintf("vault:%d", i), BCAST_ACTION_SECRET_REMOVE, nil)
		}
	}
	wg.Wait()
	for i := 0; i < 1000; i++ {
		bc.Send("team", "vault", BCAST_ACTION_SECRET_CHANGE, nil)
	}
}
