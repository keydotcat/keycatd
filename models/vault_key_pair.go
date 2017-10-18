package models

import (
	"strings"

	"github.com/keydotcat/backend/util"
)

type VaultKeyPair struct {
	PublicKey []byte            `json:"public_key"`
	Keys      map[string][]byte `json:"keys"`
}

func (vkp VaultKeyPair) checkKeysForEveryone(ppl []string) error {
	if vkp.Keys == nil {
		return util.NewErrorf("No keys for vault")
	}
	var ko []string
	for _, p := range ppl {
		if _, ok := vkp.Keys[p]; !ok {
			ko = append(ko, p)
		}
	}
	if ko != nil {
		return util.NewErrorf("Missing vault keys for %s", strings.Join(ko, ", "))
	}
	for kp := range vkp.Keys {
		found := false
		for _, p := range ppl {
			if p == kp {
				found = true
				break
			}
		}
		if !found {
			ko = append(ko, kp)
		}
	}
	if ko != nil {
		return util.NewErrorf("Extra vault keys for %s", strings.Join(ko, ", "))
	}
	return nil
}
