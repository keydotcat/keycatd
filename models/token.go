package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/keydotcat/backend/util"
)

const TOKEN_VERIFICATION = 0

type Token struct {
	Id        string    `scaneo:"pk" json:"id"`
	Type      int       `json:"-"`
	User      string    `json:"-"`
	Extra     string    `json:"extra,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func FindToken(ctx context.Context, id string) (*Token, error) {
	t := &Token{Id: id}
	err := doTx(ctx, func(tx *sql.Tx) error {
		return t.dbFind(tx)
	})
	if isNotExistsErr(err) {
		return nil, util.NewErrorFrom(ErrDoesntExist)
	}
	return t, nil
}

func FindTokensForUser(ctx context.Context, user string) (tokens []*Token) {
	doTx(ctx, func(tx *sql.Tx) error {
		tokens = findTokensForUser(tx, user)
		return nil
	})
	return tokens
}

func findTokensForUser(tx *sql.Tx, user string) []*Token {
	rows, err := tx.Query("SELECT "+selectTokenFields+" FROM \"token\" WHERE \"user\"=$1", user)
	if err != nil {
		panic(err)
	}
	ts, err := scanTokens(rows)
	if err != nil {
		panic(err)
	}
	return ts
}

func (u *Token) validate() error {
	errs := util.NewErrorFields().(*util.Error)
	if !reValidUsername.MatchString(u.User) {
		errs.SetFieldError("username", "invalid")
	}
	if len(u.Id) < 6 {
		errs.SetFieldError("id", "too short")
	}
	if u.Type != TOKEN_VERIFICATION {
		errs.SetFieldError("type", "invalid")
	}
	return errs.SetErrorOrCamo(ErrInvalidAttributes)
}

func (u *Token) insert(tx *sql.Tx) error {
	u.Id = util.GenerateRandomToken(32)
	if err := u.validate(); err != nil {
		return err
	}
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = u.CreatedAt
	_, err := u.dbInsert(tx)
	if IsDuplicateErr(err) {
		return util.NewErrorFrom(ErrAlreadyExists)
	}
	isErrOrPanic(err)
	return util.NewErrorFrom(err)
}

func (u *Token) update(tx *sql.Tx) error {
	if err := u.validate(); err != nil {
		return err
	}
	u.UpdatedAt = time.Now().UTC()
	res, err := u.dbUpdate(tx)
	return treatUpdateErr(res, err)
}

func (t *Token) ConfirmEmail(ctx context.Context) (u *User, err error) {
	if t.Type != TOKEN_VERIFICATION {
		return nil, util.NewErrorFrom(ErrDoesntExist)
	}
	return u, doTx(ctx, func(tx *sql.Tx) error {
		u, err = findUser(tx, t.User)
		if err != nil {
			return err
		}
		if err = treatUpdateErr(t.dbDelete(tx)); err != nil {
			return err
		}
		u.Email = u.UnconfirmedEmail
		u.UnconfirmedEmail = ""
		if !u.ConfirmedAt.Valid {
			u.ConfirmedAt.Valid = true
			u.ConfirmedAt.Time = time.Now().UTC()
		}
		return u.update(tx)
	})
}
