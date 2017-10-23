package models

import "time"

type Secret struct {
	Team      string    `scaneo:"pk" json:"-"`
	Vault     string    `scaneo:"pk" json:"-"`
	Id        string    `scaneo:"pk" json:"id"`
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
