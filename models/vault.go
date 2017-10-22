package models

import (
	"database/sql"
	"time"

	"github.com/keydotcat/backend/util"
)

type Vault struct {
	Id        string    `scaneo:"pk" json:"id"`
	Team      string    `scaneo:"pk" json:"team"`
	PublicKey []byte    `json:"public_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func createVault(tx *sql.Tx, id, team string, vkp VaultKeyPair) (*Vault, error) {
	v := &Vault{Id: id, Team: team, PublicKey: vkp.PublicKey}
	if err := v.insert(tx); err != nil {
		return nil, err
	}
	for u, k := range vkp.Keys {
		if err := v.addUser(tx, u, k); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (v *Vault) insert(tx *sql.Tx) error {
	if err := v.validate(); err != nil {
		return err
	}
	now := time.Now().UTC()
	v.CreatedAt = now
	v.UpdatedAt = now
	_, err := v.dbInsert(tx)
	switch {
	case isDuplicateErr(err):
		return util.NewErrorf("Vault name already exists")
	case isErrOrPanic(err):
		return err
	}
	return nil
}

func (v Vault) validate() error {
	errs := util.NewErrorFields().(*util.Error)
	if len(v.Id) == 0 {
		errs.SetFieldError("id", "missing")
	}
	if len(v.Team) == 0 {
		errs.SetFieldError("team", "missing")
	}
	if len(v.PublicKey) != 32 {
		errs.SetFieldError("public_key", "invalid")
	}
	return errs.Camo()
}

func (v Vault) addUser(tx *sql.Tx, username string, key []byte) error {
	vu := &vaultUser{Team: v.Team, Vault: v.Id, User: username, Key: key}
	return vu.insert(tx)
}
