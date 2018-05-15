package models

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/keydotcat/backend/util"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	HASH_PASSWD_COST = 14
	reValidUsername  = regexp.MustCompile(`^[\w-]{3,}$`)
	reValidEmail     = regexp.MustCompile(`^([\w-]+\.?)+@([\w-]+\.*)+\.\w+$`)
)

type User struct {
	Id               string      `scaneo:"pk" json:"id"`
	Email            string      `json:"-"`
	UnconfirmedEmail string      `json:"-"`
	HashPass         []byte      `json:"-"`
	FullName         string      `json:"fullname"`
	ConfirmedAt      pq.NullTime `json:"confirmed_at,omitempty"`
	LockedAt         pq.NullTime `json:"locked_at,omitempty"`
	SignInCount      int         `json:"sign_in_count"`
	FailedAttempts   int         `json:"failed_attempts"`
	PublicKey        []byte      `json:"public_key"`
	Key              []byte      `json:"-"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

func NewUser(ctx context.Context, id, fullname, email, password string, keyPack []byte, signedVaultKeys VaultKeyPair) (*User, *Token, error) {
	pub, priv, err := expandUserKeyPack(keyPack)
	if err != nil {
		return nil, nil, err
	}
	u := &User{
		Id:               id,
		Email:            email,
		UnconfirmedEmail: email,
		FullName:         fullname,
		PublicKey:        pub,
		Key:              priv,
	}
	vaultKeys, err := signedVaultKeys.verifyAndUnpack(u.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	t := &Token{Type: TOKEN_VERIFICATION, User: u.Id}
	if err := u.SetPassword(password); err != nil {
		return nil, nil, err
	}
	return u, t, doTx(ctx, func(tx *sql.Tx) error {
		if err := u.insert(tx); err != nil {
			return err
		}
		if err := t.insert(tx); err != nil {
			return err
		}
		_, err := createTeam(tx, u, true, u.FullName, vaultKeys)
		if err != nil {
			return err
		}
		invites, err := findInvitesForUser(tx, u.Email)
		if err != nil {
			return err
		}
		for _, i := range invites {
			team, err := i.getTeam(tx)
			if err != nil {
				return err
			}
			if err := team.addUserNoAdminCheck(tx, u); err != nil {
				return err
			}
			if _, err := i.dbDelete(tx); err != nil {
				return err
			}
		}
		return err
	})
}

func FindUser(ctx context.Context, id string) (u *User, err error) {
	return u, doTx(ctx, func(tx *sql.Tx) error {
		u, err = findUser(tx, id)
		return err
	})
}

func findUser(tx *sql.Tx, id string) (*User, error) {
	return findUserByField(tx, "id", id)
}

func FindUserByEmail(ctx context.Context, email string) (u *User, err error) {
	return u, doTx(ctx, func(tx *sql.Tx) error {
		u, err = findUserByEmail(tx, email)
		return err
	})
}

func findUserByEmail(tx *sql.Tx, email string) (*User, error) {
	return findUserByField(tx, "email", email)
}

func findUserByField(tx *sql.Tx, fieldName, value string) (*User, error) {
	r := tx.QueryRow(fmt.Sprintf(`SELECT %s FROM "user" WHERE "%s" = $1`, selectUserFields, fieldName), value)
	u := &User{}
	err := u.dbScanRow(r)
	if isNotExistsErr(err) {
		return nil, util.NewErrorFrom(ErrDoesntExist)
	}
	isErrOrPanic(err)
	return u, util.NewErrorFrom(err)
}

func (u *User) CreateTeam(ctx context.Context, name string, signedVaultKeys VaultKeyPair) (t *Team, err error) {
	vaultKeys, err := signedVaultKeys.verifyAndUnpack(u.PublicKey)
	if err != nil {
		return nil, err
	}
	return t, doTx(ctx, func(tx *sql.Tx) error {
		t, err = createTeam(tx, u, false, name, vaultKeys)
		return err
	})
}

func (u *User) insert(tx *sql.Tx) error {
	if err := u.validate(); err != nil {
		return err
	}
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = u.CreatedAt
	_, err := u.dbInsert(tx)
	if IsDuplicateErr(err) {
		dup := getDuplicateFieldFromErr(err)
		if dup == "" {
			return util.NewErrorFrom(ErrAlreadyExists)
		}
		errs := util.NewErrorFields().(*util.Error)
		errs.SetFieldError("user_"+dup, "duplicate")
		return errs.SetErrorOrCamo(ErrAlreadyExists)
	}
	isErrOrPanic(err)
	return util.NewErrorFrom(err)
}

func (u *User) update(tx *sql.Tx) error {
	if err := u.validate(); err != nil {
		return err
	}
	u.UpdatedAt = u.CreatedAt
	res, err := u.dbUpdate(tx)
	return treatUpdateErr(res, err)
}

func (u *User) validate() error {
	errs := util.NewErrorFields().(*util.Error)
	if !reValidUsername.MatchString(u.Id) {
		errs.SetFieldError("user_id", "invalid")
	}
	if len(u.FullName) == 0 {
		errs.SetFieldError("user_fullname", "invalid")
	}
	if len(u.HashPass) < 6 {
		errs.SetFieldError("user_password", "too short")
	}
	if !reValidEmail.MatchString(u.Email) {
		errs.SetFieldError("user_email", "invalid")
	}
	if len(u.PublicKey) != publicKeyPackSize {
		errs.SetFieldError("user_public_key", "invalid")
	}
	if len(u.Key) != privateKeyPackSize {
		errs.SetFieldError("user_private_key", "invalid")
	}
	return errs.SetErrorOrCamo(ErrAlreadyExists)
}

func (u *User) SetPassword(pass string) error {
	hpas, err := bcrypt.GenerateFromPassword([]byte(pass), HASH_PASSWD_COST)
	if err != nil {
		panic(err)
	}
	u.HashPass = hpas
	return nil
}

func (u *User) CheckPassword(pass string) error {
	err := bcrypt.CompareHashAndPassword(u.HashPass, []byte(pass))
	if err != nil {
		return util.NewErrorFrom(ErrUnauthorized)
	}
	return nil
}

func (u *User) GetTeams(ctx context.Context) ([]*Team, error) {
	db := GetDB(ctx)
	rows, err := db.Query(`SELECT `+selectTeamFullFields+` FROM "team", "team_user" WHERE  "team_user"."team" = "team".id AND "team_user"."user" = $1`, u.Id)
	if isErrOrPanic(err) {
		return nil, util.NewErrorFrom(err)
	}
	teams, err := scanTeams(rows)
	isErrOrPanic(err)
	return teams, util.NewErrorFrom(err)
}

func (u *User) GetVerificationToken(ctx context.Context) (t *Token, err error) {
	t = &Token{Id: u.Id, Type: TOKEN_VERIFICATION}
	err = doTx(ctx, func(tx *sql.Tx) error {
		r := tx.QueryRow(`SELECT `+selectTokenFields+` FROM "token" WHERE "user" = $1 AND "type" = $2"`, u.Id, TOKEN_VERIFICATION)
		return t.dbScanRow(r)
	})
	if isNotExistsErr(err) {
		return nil, util.NewErrorFrom(ErrDoesntExist)
	}
	return t, nil
}

func (u *User) GetTeam(ctx context.Context, tid string) (t *Team, err error) {
	return t, doTx(ctx, func(tx *sql.Tx) error {
		t, err = u.getTeam(tx, tid)
		return err
	})
}

func (u *User) getTeam(tx *sql.Tx, tid string) (*Team, error) {
	t := &Team{}
	r := tx.QueryRow(`SELECT `+selectTeamFullFields+` FROM "team", "team_user" WHERE "team_user"."user" = $1 AND "team_user"."team" = "team"."id" AND "team"."id" = $2`, u.Id, tid)
	err := t.dbScanRow(r)
	if isNotExistsErr(err) {
		return nil, util.NewErrorFrom(ErrDoesntExist)
	}
	isErrOrPanic(err)
	return t, err
}
