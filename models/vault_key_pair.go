package models

import (
	"github.com/keydotcat/backend/util"
	"golang.org/x/crypto/ed25519"
)

const publicKeyPackSize = ed25519.PublicKeySize + 32
const privateKeyPackMinSize = ed25519.PrivateKeySize + 32

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

func expandUserKeyPack(pack []byte) ([]byte, []byte, error) {
	if len(pack) < ed25519.PublicKeySize+ed25519.PrivateKeySize+ed25519.SignatureSize+64 {
		return nil, nil, util.NewErrorFrom(ErrInvalidKeys)
	}
	spub := pack[:ed25519.PublicKeySize]
	verifiedPack, err := verifyAndUnpack(spub, pack[ed25519.PublicKeySize:])
	if err != nil {
		return nil, nil, util.NewErrorFrom(ErrInvalidSignature)
	}
	return append(spub, verifiedPack[:32]...), verifiedPack[32:], nil
}

func verifyAndUnpack(pubPack []byte, signedData []byte) ([]byte, error) {
	pubkey := pubPack[:ed25519.PublicKeySize]
	if l := len(pubkey); l != ed25519.PublicKeySize {
		return nil, util.NewErrorFrom(ErrInvalidPublicKey)
	}
	if len(signedData) < ed25519.SignatureSize || signedData[63]&224 != 0 {
		return nil, util.NewErrorFrom(ErrInvalidSignature)
	}
	sign := signedData[:ed25519.SignatureSize]
	msg := signedData[ed25519.SignatureSize:]

	if ed25519.Verify(ed25519.PublicKey(pubkey), msg, sign) {
		return msg, nil
	}
	return nil, util.NewErrorFrom(ErrInvalidSignature)
}

func (vkp VaultKeyPair) verifyAndUnpack(pubkey []byte) (VaultKeyPair, error) {
	unpacked := VaultKeyPair{Keys: map[string][]byte{}}
	if data, err := verifyAndUnpack(pubkey, vkp.PublicKey); err != nil {
		return unpacked, err
	} else {
		unpacked.PublicKey = data
	}
	for k, v := range vkp.Keys {
		if data, err := verifyAndUnpack(unpacked.PublicKey, v); err != nil {
			return unpacked, err
		} else {
			unpacked.Keys[k] = data
		}
	}
	return unpacked, nil
}
