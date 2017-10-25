package models

import (
	"bytes"
	"crypto/rand"
	"testing"

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

func getDummyVaultKeyPair(signerPack []byte, ids ...string) VaultKeyPair {
	pubPack, privPack, _ := generateNewKeys()
	signer := signerPack[:ed25519.PrivateKeySize]
	vkp := VaultKeyPair{signAndPack(signer, pubPack), map[string][]byte{}}
	for _, id := range ids {
		vkp.Keys[id] = signAndPack(privPack[:ed25519.PrivateKeySize], privPack)
	}
	return vkp
}

func expandVaultKeysOnce(vs []*VaultFull) VaultKeyPair {
	vkp := VaultKeyPair{Keys: map[string][]byte{}}
	for _, v := range vs {
		vkp.Keys[v.Id] = signAndPack(v.Key[:ed25519.PrivateKeySize], v.Key)
	}
	return vkp
}

func TestSigningAndVerificationOfVaultKeyPairs(t *testing.T) {
	pub, priv, fullPack := generateNewKeys()
	epub, epriv, err := expandUserKeyPack(fullPack)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(pub, epub) != 0 || bytes.Compare(priv, epriv) != 0 {
		t.Fatalf("Packed and unpacked keys differ")
	}
	vkp := getDummyVaultKeyPair(priv, "random1", "random2")
	unpacked, err := vkp.verifyAndUnpack(pub)
	if err != nil {
		t.Fatal(err)
	}
	vkp = getDummyVaultKeyPair(unpacked.Keys["random1"])
	if _, err := vkp.verifyAndUnpack(unpacked.PublicKey); err != nil {
		t.Fatal(err)
	}
}
