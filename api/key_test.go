package api

import (
	"crypto/rand"

	"github.com/keydotcat/server/models"
	"github.com/keydotcat/server/util"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

var boxNonceSize = 24
var a32b = make([]byte, 32)

func generateNewKeys() ([]byte, []byte, []byte) {
	epub, epriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	spub, spriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	pubPack := append([]byte(spub), signAndPack(spriv, (*epub)[:])...)
	priv := append([]byte(spriv), (*epriv)[:]...)
	snonce := util.GenerateRandomByteArray(boxNonceSize)
	var nonce [24]byte
	copy(nonce[:], snonce)
	var spass [32]byte
	copy(spass[:], spub)
	privPack := signAndPack(spriv, secretbox.Seal(snonce, priv, &nonce, &spass))
	return pubPack, priv, append(pubPack, privPack...)
}

func signAndPack(keyPack []byte, msg []byte) []byte {
	key := keyPack[:ed25519.PrivateKeySize]
	sig := ed25519.Sign(ed25519.PrivateKey(key), msg)
	response := make([]byte, ed25519.SignatureSize+len(msg))
	copy(response[:ed25519.SignatureSize], sig)
	copy(response[ed25519.SignatureSize:], msg)
	return response
}

func getDummyVaultKeyPair(signerPack []byte, ids ...string) models.VaultKeyPair {
	pubPack, privPack, _ := generateNewKeys()
	adminSigner := signerPack[:ed25519.PrivateKeySize]
	vaultSigner := privPack[:ed25519.PrivateKeySize]
	vkp := models.VaultKeyPair{
		PublicKey: signAndPack(adminSigner, pubPack),
		Keys:      map[string][]byte{},
	}
	snonce := util.GenerateRandomByteArray(boxNonceSize)
	var nonce [24]byte
	copy(nonce[:], snonce)
	var sharedK [32]byte
	copy(sharedK[:], pubPack)
	signedsealed := signAndPack(vaultSigner, box.SealAfterPrecomputation(snonce, privPack, &nonce, &sharedK))
	for _, id := range ids {
		vkp.Keys[id] = signedsealed
	}
	return vkp
}

func unsealVaultKey(v *models.Vault, signedsealed []byte) []byte {
	var sharedK [32]byte
	copy(sharedK[:], v.PublicKey)
	sealed, err := verifyAndUnpack(v.PublicKey, signedsealed)
	if err != nil {
		panic(err)
	}
	var nonce [24]byte
	copy(nonce[:], sealed[:24])
	unsealed, ok := box.OpenAfterPrecomputation(nil, sealed[24:], &nonce, &sharedK)
	if !ok {
		panic("Could not unseal box")
	}
	return unsealed
}

func expandVaultKeysOnce(vs []*models.VaultFull) models.VaultKeyPair {
	vkp := models.VaultKeyPair{Keys: map[string][]byte{}}
	for _, v := range vs {
		vkp.Keys[v.Id] = v.Key
	}
	return vkp
}

func verifyAndUnpack(pubPack []byte, signedData []byte) ([]byte, error) {
	pubKey := pubPack[:ed25519.PublicKeySize]
	if l := len(pubKey); l != ed25519.PublicKeySize {
		return nil, util.NewErrorFrom(models.ErrInvalidPublicKey)
	}
	if len(signedData) < ed25519.SignatureSize || signedData[63]&224 != 0 {
		return nil, util.NewErrorFrom(models.ErrInvalidSignature)
	}
	sign := signedData[:ed25519.SignatureSize]
	msg := signedData[ed25519.SignatureSize:]

	if ed25519.Verify(ed25519.PublicKey(pubKey), msg, sign) {
		return msg, nil
	}
	return nil, util.NewErrorFrom(models.ErrInvalidSignature)
}

func getUserPrivateKeys(pubPack, privPack []byte) []byte {
	closed, err := verifyAndUnpack(pubPack, privPack)
	if err != nil {
		panic(err)
	}
	var nonce [24]byte
	copy(nonce[:], closed)
	closed = closed[24:]
	var spass [32]byte
	copy(spass[:], pubPack)
	open, ok := secretbox.Open([]byte{}, closed, &nonce, &spass)
	if !ok {
		panic("Could not open secret box")
	}
	return open
}

func getVaultPrivateKeys(pubPack, privPack []byte) []byte {
	closed, err := verifyAndUnpack(pubPack, privPack)
	if err != nil {
		panic(err)
	}
	var nonce [24]byte
	copy(nonce[:], closed)
	closed = closed[24:]
	var spass [32]byte
	copy(spass[:], pubPack)
	open, ok := box.OpenAfterPrecomputation([]byte{}, closed, &nonce, &spass)
	if !ok {
		panic("Could not open  box")
	}
	return open
}
