package models

import "github.com/keydotcat/backend/util"

type VaultKeyPair struct {
	PublicKey []byte            `json:"public_key"`
	Keys      map[string][]byte `json:"keys"`
}

func (vkp VaultKeyPair) checkKeyIdsMatch(ppl []string) error {
	if vkp.Keys == nil {
		return util.NewErrorFrom(ErrInvalidKeys)
	}
	var ko []string
	for _, p := range ppl {
		if _, ok := vkp.Keys[p]; !ok {
			ko = append(ko, p)
		}
	}
	if ko != nil {
		return util.NewErrorFrom(ErrInvalidKeys)
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
		return util.NewErrorFrom(ErrInvalidKeys)
	}
	return nil
}
