package models

import "errors"

var (
	ErrInvalidEmail      = errors.New("Invalid email")
	ErrNotInTeam         = errors.New("User does not belong to team")
	ErrUnauthorized      = errors.New("You cannot do that")
	ErrAlreadyInTeam     = errors.New("Already belongs to team")
	ErrAlreadyInvited    = errors.New("Alredy invited")
	ErrAlreadyExists     = errors.New("Already exists")
	ErrInvalidKeys       = errors.New("Invalid keys for vault")
	ErrDoesntExist       = errors.New("Does not exist")
	ErrInvalidSignature  = errors.New("Invalid signature")
	ErrInvalidPublicKey  = errors.New("Invalid public key length")
	ErrInvalidAttributes = errors.New("Invalid attributes")
)
