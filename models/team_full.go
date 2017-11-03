package models

import "context"

type TeamFull struct {
	*Team
	Vaults []*VaultFull `json:"vaults"`
}

func (t *Team) GetTeamFull(ctx context.Context, u *User) (*TeamFull, error) {
	vf, err := t.GetVaultsFullForUser(ctx, u)
	if err != nil {
		return nil, err
	}
	return &TeamFull{t, vf}, nil
}
