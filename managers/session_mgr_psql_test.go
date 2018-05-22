package managers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"testing"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/keydotcat/backend/db"
	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

var mdb *sql.DB

func init() {
	var err error
	mdb, err = sql.Open("postgres", "user=root dbname=test sslmode=disable port=26257")
	if err != nil {
		panic(err)
	}
	m := db.NewMigrateMgr(mdb)
	if err := m.LoadMigrations(); err != nil {
		panic(err)
	}
	lid, ap, err := m.ApplyRequiredMigrations()
	if err != nil {
		panic(err)
	}
	log.Printf("Executed migrations until %d (%d applied)", lid, ap)
}

func getCtx() context.Context {
	return models.AddDBToContext(context.Background(), mdb)
}

const boxNonceSize = 24
const boxPublicKeySize = 32
const boxPrivateKeySize = 32

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

func getDummyVaultKeyPair(signerPack []byte, ids ...string) models.VaultKeyPair {
	pubPack, privPack, _ := generateNewKeys()
	adminSigner := signerPack[:ed25519.PrivateKeySize]
	vaultSigner := privPack[:ed25519.PrivateKeySize]
	vkp := models.VaultKeyPair{signAndPack(adminSigner, pubPack), map[string][]byte{}}
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

func signAndPack(keyPack []byte, msg []byte) []byte {
	key := keyPack[:ed25519.PrivateKeySize]
	sig := ed25519.Sign(ed25519.PrivateKey(key), msg)
	response := make([]byte, ed25519.SignatureSize+len(msg))
	copy(response[:ed25519.SignatureSize], sig)
	copy(response[ed25519.SignatureSize:], msg)
	return response
}

func getDummyUser() *models.User {
	ctx := getCtx()
	uid := "u_" + util.GenerateRandomToken(10)
	_, priv, fullpack := generateNewKeys()
	vkp := getDummyVaultKeyPair(priv, uid)
	u, _, err := models.NewUser(ctx, uid, "uid fullname", uid+"@nowhere.net", uid, fullpack, vkp)
	if err != nil {
		panic(err)
	}
	return u
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

func addSessionForUser(rs SessionMgr, uid, agent string) error {
	s, err := rs.NewSession(uid, agent, false)
	if err != nil {
		return err
	}
	if s.UserId != uid {
		return fmt.Errorf("Mismatch in the user id: %s vs %s", s.UserId, uid)
	}
	return nil
}

func testSessionManager(rs SessionMgr, t *testing.T, smName string) {
	uid1 := getDummyUser().Id
	uid2 := getDummyUser().Id
	if err := addSessionForUser(rs, uid1, uid1+":s1"); err != nil {
		t.Fatalf("%s failed test: %s", smName, err)
	}
	if err := addSessionForUser(rs, uid1, uid1+":s2"); err != nil {
		t.Fatalf("%s failed test: %s", smName, err)
	}
	if err := addSessionForUser(rs, uid2, uid2+":s1"); err != nil {
		t.Fatalf("%s failed test: %s", smName, err)
	}
	sess, err := rs.GetAllSessions(uid1)
	if err != nil {
		t.Fatalf("%s failed test: %s", smName, err)
	}
	sids := map[string]bool{uid1 + ":s1": false, uid1 + ":s2": false}
	for _, ses := range sess {
		sids[ses.Agent] = true
	}
	for k, v := range sids {
		if !v {
			t.Errorf("%s didn't find session %s", smName, k)
		}
	}
	s, err := rs.UpdateSession(sess[0].Id, sess[0].Agent+":")
	if err != nil {
		t.Fatalf("%s failed test: %s", smName, err)
	}
	if s.Id != sess[0].Id || s.Agent != sess[0].Agent+":" {
		t.Fatalf("%s session hasn't been updated", smName)
	}
	_, err = rs.UpdateSession("nonexistant", "asd")
	if !util.CheckErr(err, models.ErrDoesntExist) {
		t.Fatalf("%s Unexpected error: %s vs %s", smName, models.ErrDoesntExist, err)
	}
	if err = rs.DeleteSession(sess[0].Id); err != nil {
		t.Fatalf("%s failed test: %s", smName, err)
	}
	_, err = rs.UpdateSession(sess[0].Id, "asd")
	if !util.CheckErr(err, models.ErrDoesntExist) {
		t.Fatalf("%s unexpected error: %s vs %s", smName, models.ErrDoesntExist, err)
	}
	for _, uid := range []string{uid1, uid2} {
		if err = rs.DeleteAllSessions(uid1); err != nil {
			t.Fatalf("%s could not delete all sessions for user %s: %s", smName, uid, err)
		}
	}
}
func TestPSQLSessionManager(t *testing.T) {
	mdb, err := sql.Open("postgres", "user=root dbname=test sslmode=disable port=26257")
	if err != nil {
		t.Fatal(err)
	}
	rs := NewSessionMgrPSQL(mdb)
	testSessionManager(rs, t, "psql")
}
