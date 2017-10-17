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
	Id               string
	Email            string
	UnconfirmedEmail string
	HashedPassword   []byte
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

func NewUser(ctx context.Context, id, fullname, email, password string, pubkey, key []byte) (*User, error) {
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
		_, err := createUserTeam(tx, u, u.FullName)
		return err
	})
}

func (u *User) insert(tx *sql.Tx) error {
	if err := u.validate(); err != nil {
		return err
	}
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = u.CreatedAt
	_, err := tx.Exec(fmt.Sprintf(`INSERT INTO "Users" %s VALUES %s`, insertUserFields, insertUserBind), fieldsUser(u)...)
	if err != nil {
		return util.NewErrorf("Could not create user: %s", err)
	}
	return nil
}

func (u *User) validate() error {
	errs := util.NewErrorFields().(*util.Error)
	if !reValidUsername.MatchString(u.Id) {
		errs.SetFieldError("username", "invalid")
	}
	if len(u.FullName) == 0 {
		errs.SetFieldError("fullname", "invalid")
	}
	if len(u.HashedPassword) < 6 {
		errs.SetFieldError("password", "too short")
	}
	if !reValidEmail.MatchString(u.Email) {
		errs.SetFieldError("email", "invalid")
	}
	if len(u.PublicKey) != 32 {
		errs.SetFieldError("public_key", "invalid")
	}
	if len(u.Key) != 32 {
		errs.SetFieldError("private_key", "invalid")
	}
	return errs.Camo()
}

func (u *User) SetPassword(pass string) error {
	hpas, err := bcrypt.GenerateFromPassword([]byte(pass), HASH_PASSWD_COST)
	if err != nil {
		return util.NewErrorf("Could not hash password: %s", err)
	}
	u.HashedPassword = hpas
	return nil
}

func (u *User) CheckPassword(pass string) error {
	return bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(pass))
}
