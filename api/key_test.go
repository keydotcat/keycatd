package api

import (
	"crypto/rand"

	"github.com/keydotcat/backend/models"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
)

func generateNewKeys() ([]byte, []byte, []byte) {
	epub, epriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	spub, spriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	pub := append([]byte(spub), (*epub)[:]...)
	priv := append([]byte(spriv), (*epriv)[:]...)
	return pub, priv, append([]byte(spub), signAndPack(priv, append((*epub)[:], priv...))...)
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
	signer := signerPack[:ed25519.PrivateKeySize]
	vkp := models.VaultKeyPair{signAndPack(signer, pubPack), map[string][]byte{}}
	for _, id := range ids {
		vkp.Keys[id] = signAndPack(privPack[:ed25519.PrivateKeySize], privPack)
	}
	return vkp
}

func expandVaultKeysOnce(vs []*models.VaultFull) models.VaultKeyPair {
	vkp := models.VaultKeyPair{Keys: map[string][]byte{}}
	for _, v := range vs {
		vkp.Keys[v.Id] = signAndPack(v.Key[:ed25519.PrivateKeySize], v.Key)
	}
	return vkp
}
