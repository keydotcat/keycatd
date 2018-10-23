package models

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/keydotcat/keycatd/util"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
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
	adminSigner := signerPack[:ed25519.PrivateKeySize]
	vaultSigner := privPack[:ed25519.PrivateKeySize]
	vkp := VaultKeyPair{signAndPack(adminSigner, pubPack), map[string][]byte{}}
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

func unsealVaultKey(v *Vault, signedsealed []byte) []byte {
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

func expandVaultKeysOnce(vs []*VaultFull) VaultKeyPair {
	vkp := VaultKeyPair{Keys: map[string][]byte{}}
	for _, v := range vs {
		vkp.Keys[v.Id] = v.Key
	}
	return vkp
}

func TestSigningAndVerificationOfVaultKeyPairs(t *testing.T) {
	pub, priv, fullPack := generateNewKeys()
	epub, epriv, err := expandUserKeyPack(fullPack)
	if err != nil {
		t.Fatal(err)
	}
	openPrivKeys := getUserPrivateKeys(epub, epriv)
	if bytes.Compare(pub, epub) != 0 || bytes.Compare(priv, openPrivKeys) != 0 {
		t.Fatalf("Packed and unpacked keys differ")
	}
	vkp := getDummyVaultKeyPair(priv, "random1", "random2")
	unpacked, err := vkp.verifyAndUnpack(pub)
	if err != nil {
		t.Fatal(err)
	}
	vkp = getDummyVaultKeyPair(getVaultPrivateKeys(unpacked.PublicKey, unpacked.Keys["random1"]))
	if _, err := vkp.verifyAndUnpack(unpacked.PublicKey); err != nil {
		t.Fatal(err)
	}
}
