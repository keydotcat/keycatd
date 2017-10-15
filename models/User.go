package models

import (
	"context"

	"github.com/lib/pq"
)

type User struct {
	Id               string
	Email            string
	UnconfirmedEmail string
	Password         string
	FullName         string
	ConfirmedAt      pq.NullTime
	LockedAt         pq.NullTime
	SignInCount      int
	FailedAttempts   int
	PublicKey        []byte
	Key              []byte
}

func NewUser(ctx context.Context, id, email, password, fullname string, pubkey, key []byte) (*User, error) {
	return nil, nil
}
