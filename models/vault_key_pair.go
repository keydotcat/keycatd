package models

import (
	"github.com/keydotcat/keycatd/util"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/secretbox"
)

const boxNonceSize = 24
const boxPublicKeySize = 32
const boxPrivateKeySize = 32

var (
	publicKeyPackSize  = ed25519.PublicKeySize + ed25519.SignatureSize + boxPublicKeySize
	privateKeyPackSize = ed25519.SignatureSize + secretbox.Overhead + boxNonceSize + ed25519.PrivateKeySize + boxPrivateKeySize
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

func expandUserKeyPack(pack []byte) ([]byte, []byte, error) {
	if len(pack) < publicKeyPackSize+privateKeyPackSize {
		return nil, nil, util.NewErrorFrom(ErrInvalidKeys)
	}
	pubPack := pack[:publicKeyPackSize]
	privPack := pack[publicKeyPackSize:]
	spub := pubPack[:ed25519.PublicKeySize]
	_, err := verifyAndUnpack(spub, pubPack[ed25519.PublicKeySize:])
	if err != nil {
		return nil, nil, util.NewErrorFrom(ErrInvalidSignature)
	}
	_, err = verifyAndUnpack(spub, privPack)
	if err != nil {
		return nil, nil, util.NewErrorFrom(ErrInvalidSignature)
	}
	return pubPack, privPack, nil
}

func verifyAndUnpack(pubPack []byte, signedData []byte) ([]byte, error) {
	pubKey := pubPack[:ed25519.PublicKeySize]
	if l := len(pubKey); l != ed25519.PublicKeySize {
		return nil, util.NewErrorFrom(ErrInvalidPublicKey)
	}
	if len(signedData) < ed25519.SignatureSize || signedData[63]&224 != 0 {
		return nil, util.NewErrorFrom(ErrInvalidSignature)
	}
	sign := signedData[:ed25519.SignatureSize]
	msg := signedData[ed25519.SignatureSize:]

	if ed25519.Verify(ed25519.PublicKey(pubKey), msg, sign) {
		return msg, nil
	}
	return nil, util.NewErrorFrom(ErrInvalidSignature)
}

func (vkp VaultKeyPair) verifyAndUnpack(pubPack []byte) (VaultKeyPair, error) {
	unpacked := VaultKeyPair{Keys: map[string][]byte{}}
	if data, err := verifyAndUnpack(pubPack, vkp.PublicKey); err != nil {
		return unpacked, err
	} else {
		unpacked.PublicKey = data
	}
	for k, v := range vkp.Keys {
		if _, err := verifyAndUnpack(unpacked.PublicKey, v); err != nil {
			return unpacked, err
		} else {
			unpacked.Keys[k] = v
		}
	}
	return unpacked, nil
}
