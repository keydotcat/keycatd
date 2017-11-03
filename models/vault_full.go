package models

import (
	"context"
	"database/sql"

	"github.com/keydotcat/backend/util"
)

type VaultFull struct {
	Vault
	Key []byte `json:"key"`
}

func scanVaultsFull(rs *sql.Rows) ([]*VaultFull, error) {
	structs := make([]*VaultFull, 0, 16)
	var err error
	for rs.Next() {
		var s VaultFull
		if err = rs.Scan(
			&s.Id,
			&s.Team,
			&s.PublicKey,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.Key,
		); err != nil {
			return nil, err
		}
		structs = append(structs, &s)
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}

func (t *Team) GetVaultsFullForUser(ctx context.Context, u *User) (vf []*VaultFull, err error) {
	return vf, doTx(ctx, func(tx *sql.Tx) error {
		vf, err = t.getVaultsFullForUser(tx, u)
		return err
	})
}

func (t *Team) getVaultsFullForUser(tx *sql.Tx, u *User) ([]*VaultFull, error) {
	rows, err := tx.Query(`SELECT `+selectVaultFullFields+`, "vault_user"."key" FROM "vault", "vault_user" WHERE  "vault"."team" = $1 AND "vault"."team" = "vault_user"."team" AND "vault"."id" = "vault_user"."vault" AND "vault_user"."user" = $2`, t.Id, u.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	vaults, err := scanVaultsFull(rows)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	return vaults, nil
}
