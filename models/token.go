package models

type Token struct {
	Id   string `scaneo:"pk"`
	Type string
	User string
}
