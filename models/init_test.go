package models

import (
	"context"
	"database/sql"
	"log"

	"github.com/keydotcat/keycatd/db"
	"github.com/keydotcat/keycatd/thelpers"
	"github.com/keydotcat/keycatd/util"
	"golang.org/x/crypto/bcrypt"
)

var mdb *sql.DB

var DUMP_BEFORE_TEST = false

func init() {
	HASH_PASSWD_COST = bcrypt.MinCost
	mdb = thelpers.GetDBConn()
	thelpers.DropAllTables(mdb)
	m := db.NewMigrateMgr(mdb, thelpers.GetTestDBType())
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
	return AddDBToContext(context.Background(), mdb)
}

type vaultMock struct {
	v    *Vault
	priv []byte
}

func createVaultMock(user *User, team *Team) vaultMock {
	ctx := getCtx()
	userPrivKeys := getUserPrivateKeys(user.PublicKey, user.Key)
	vkp := getDummyVaultKeyPair(userPrivKeys, user.Id)
	v, err := team.CreateVault(ctx, user, util.GenerateRandomToken(5), vkp)
	if err != nil {
		panic(err)
	}
	return vaultMock{v, unsealVaultKey(v, vkp.Keys[user.Id])}
}

func getFirstVault(o *User, t *Team) vaultMock {
	vs, err := t.GetVaultsFullForUser(getCtx(), o)
	if err != nil {
		panic(err)
	}
	return vaultMock{&vs[0].Vault, unsealVaultKey(&vs[0].Vault, vs[0].Key)}
}

func createTeamMock(user *User) *Team {
	ctx := getCtx()
	privKeys := getUserPrivateKeys(user.PublicKey, user.Key)
	vkp := getDummyVaultKeyPair(privKeys, user.Id)
	team, err := user.CreateTeam(ctx, "team:"+util.GenerateRandomToken(10), vkp)
	if err != nil {
		panic(nil)
	}
	return team
}
