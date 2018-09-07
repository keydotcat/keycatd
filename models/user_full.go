package models

import (
	"context"
	"fmt"

	"github.com/keydotcat/server/util"
)

type UserFull struct {
	*User
	Teams []*Team `json:"teams"`
}

func (u *User) GetUserFull(ctx context.Context) (*UserFull, error) {
	cmd := fmt.Sprintf(`SELECT %s FROM "team", "team_user" WHERE "team"."id" = "team_user"."team" AND "team_user"."user" = $1`, selectTeamFullFields)
	rows, err := GetDB(ctx).Query(cmd, u.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	teams, err := scanTeams(rows)
	isErrOrPanic(err)
	return &UserFull{u, teams}, nil
}
