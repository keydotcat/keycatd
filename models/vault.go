package models

import "time"

type Vault struct {
	Id        string    `scaneo:"pk" json:"id"`
	Team      string    `scaneo:"pk" json:"team"`
	PublicKey []byte    `json:"public_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
