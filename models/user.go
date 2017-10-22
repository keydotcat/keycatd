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
	HASH_PASSWD_COST = bcrypt.MaxCost
	reValidUsername  = regexp.MustCompile(`^[\w-]{3,}$`)
	reValidEmail     = regexp.MustCompile(`^([\w-]+\.?)+@([\w-]+\.*)+\.\w+$`)
)

type User struct {
	Id               string `scaneo:"pk" json:"id"`
	Email            string
	UnconfirmedEmail string
	HashPass         []byte
	FullName         string
	ConfirmedAt      pq.NullTime
	LockedAt         pq.NullTime
	SignInCount      int
	FailedAttempts   int
	PublicKey        []byte
	Key              []byte
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewUser(ctx context.Context, id, fullname, email, password string, pubkey, key []byte, vaultKeys VaultKeyPair) (*User, error) {
	u := &User{
		Id:               id,
		Email:            email,
		UnconfirmedEmail: email,
		FullName:         fullname,
		PublicKey:        pubkey,
		Key:              key,
	}
	if err := u.SetPassword(password); err != nil {
		return nil, err
	}
	return u, doTx(ctx, func(tx *sql.Tx) error {
		if err := u.insert(tx); err != nil {
			return err
		}
		_, err := createTeam(tx, u, true, u.FullName, vaultKeys)
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
		return nil, nil
	}
	isErrOrPanic(err)
	return u, util.NewErrorFrom(err)
}

func (u *User) CreateTeam(ctx context.Context, name string, vaultKeys VaultKeyPair) (t *Team, err error) {
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
	if isDuplicateErr(err) {
		return util.NewErrorf("Username already taken")
	}
	isErrOrPanic(err)
	return util.NewErrorFrom(err)
}

func (u *User) update(tx *sql.Tx) error {
	if err := u.validate(); err != nil {
		return err
	}
	u.UpdatedAt = u.CreatedAt
	_, err := u.dbUpdate(tx)
	isErrOrPanic(err)
	return util.NewErrorFrom(err)
}

func (u *User) validate() error {
	errs := util.NewErrorFields().(*util.Error)
	if !reValidUsername.MatchString(u.Id) {
		errs.SetFieldError("username", "invalid")
	}
	if len(u.FullName) == 0 {
		errs.SetFieldError("fullname", "invalid")
	}
	if len(u.HashPass) < 6 {
		errs.SetFieldError("password", "too short")
	}
	if !reValidEmail.MatchString(u.Email) {
		errs.SetFieldError("email", "invalid")
	}
	if len(u.PublicKey) != 32 {
		errs.SetFieldError("public_key", "invalid")
	}
	if len(u.Key) < 32 {
		errs.SetFieldError("private_key", "invalid")
	}
	return errs.Camo()
}

func (u *User) SetPassword(pass string) error {
	hpas, err := bcrypt.GenerateFromPassword([]byte(pass), HASH_PASSWD_COST)
	if err != nil {
		return util.NewErrorf("Could not hash password: %s", err)
	}
	u.HashPass = hpas
	return nil
}

func (u *User) CheckPassword(pass string) error {
	return bcrypt.CompareHashAndPassword(u.HashPass, []byte(pass))
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
