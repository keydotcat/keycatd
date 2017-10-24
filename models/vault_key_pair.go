package models

import (
	"github.com/keydotcat/backend/util"
	"golang.org/x/crypto/ed25519"
)

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

const (
	PublicKeySize = 32
	SignatureSize = 64
)

func verifyAndUnpack(pubkey []byte, signedData []byte) ([]byte, error) {
	if l := len(pubkey); l != PublicKeySize {
		return nil, util.NewErrorFrom(ErrInvalidPublicKey)
	}
	if len(signedData) < SignatureSize || signedData[63]&224 != 0 {
		return nil, util.NewErrorFrom(ErrInvalidSignature)
	}
	msg := signedData[SignatureSize:]
	sign := signedData[:SignatureSize]

	if ed25519.Verify(ed25519.PublicKey(pubkey), sign, msg) {
		return msg, nil
	}
	return nil, util.NewErrorFrom(ErrInvalidSignature)
}

func (vkp VaultKeyPair) verifyAndUnpack(pubkey []byte) (VaultKeyPair, error) {
	unpacked := VaultKeyPair{}
	if data, err := verifyAndUnpack(pubkey, vkp.PublicKey); err != nil {
		return unpacked, err
	} else {
		unpacked.PublicKey = data
	}
	for k, v := range vkp.Keys {
		if data, err := verifyAndUnpack(pubkey, v); err != nil {
			return unpacked, err
		} else {
			unpacked.Keys[k] = data
		}
	}
	return unpacked, nil
}
