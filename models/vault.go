package models

import "time"

type Vault struct {
	Id        string    `sql:"id"`
	Team      string    `sql:"team"`
	PublicKey []byte    `sql:"public_key"`
	CreatedAt time.Time `sql:"created_at"`
	UpdatedAt time.Time `sql:"updated_at"`
}
