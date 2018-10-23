package models

import (
	"context"
	"database/sql"

	"github.com/keydotcat/keycatd/util"
)

type TeamUserFull struct {
	Team           string `scaneo:"pk" json:"-"`
	User           string `scaneo:"pk" json:"id"`
	Admin          bool   `json:"admin"`
	AccessRequired bool   `json:"-"`
	FullName       string `json:"fullname"`
	PublicKey      []byte `json:"public_key"`
}

func scanTeamUserFull(rs *sql.Rows) ([]*TeamUserFull, error) {
	structs := make([]*TeamUserFull, 0, 16)
	var err error
	for rs.Next() {
		var s TeamUserFull
		if err = rs.Scan(
			&s.Team,
			&s.User,
			&s.Admin,
			&s.AccessRequired,
			&s.FullName,
			&s.PublicKey,
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

func (t *Team) GetUsersAfiliationFull(ctx context.Context) (tuf []*TeamUserFull, err error) {
	return tuf, doTx(ctx, func(tx *sql.Tx) error {
		tuf, err = t.getUsersAfiliationFull(tx)
		return err
	})
}

func (t *Team) getUsersAfiliationFull(tx *sql.Tx) ([]*TeamUserFull, error) {
	rows, err := tx.Query(`SELECT `+selectTeamUserFullFields+`, "user"."full_name", "user"."public_key" FROM "team_user", "user" WHERE "team_user"."team" = $1 AND "team_user"."user" = "user"."id"`, t.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	users, err := scanTeamUserFull(rows)
	isErrOrPanic(err)
	return users, util.NewErrorFrom(err)
}
