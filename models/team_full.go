package models

import (
	"context"
	"database/sql"
)

type TeamFull struct {
	*Team
	Vaults []*VaultFull `json:"vaults"`
	Users  []*teamUser  `json:"users"`
}

func (u *User) GetTeamFull(ctx context.Context, tid string) (tf *TeamFull, err error) {
	tf = &TeamFull{}
	return tf, doTx(ctx, func(tx *sql.Tx) error {
		t, err := u.getTeam(tx, tid)
		if err != nil {
			return err
		}
		tf, err = t.getTeamFull(tx, u)
		return err
	})
}

func (t *Team) GetTeamFull(ctx context.Context, u *User) (tf *TeamFull, err error) {
	return tf, doTx(ctx, func(tx *sql.Tx) error {
		tf, err = t.getTeamFull(tx, u)
		return err
	})
}

func (t *Team) getTeamFull(tx *sql.Tx, u *User) (*TeamFull, error) {
	vf, err := t.getVaultsFullForUser(tx, u)
	if err != nil {
		return nil, err
	}
	tu, err := t.getUsersAfiliation(tx)
	return &TeamFull{t, vf, tu}, nil
}
