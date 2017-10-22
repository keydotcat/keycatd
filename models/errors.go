package models

import "errors"

var (
	ErrInvalidEmail   = errors.New("Invalid email")
	ErrNotInTeam      = errors.New("User does not belong to team")
	ErrUnauthorized   = errors.New("You can't do that")
	ErrAlreadyInTeam  = errors.New("Already belongs to team")
	ErrAlreadyInvited = errors.New("Alredy invited")
)
