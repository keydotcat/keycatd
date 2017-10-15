package models

type User struct {
 Id string
 Email string
 UnconfirmedEmail string
 Password string
 FullName string
 ConfirmedAt time.Time
 LockedAt time.Time
 SignInCount int
 FailedAttempts int
 PublicKey []byte
 Key []byte
}


func NewUser(ctx context.Context, id, email, password, fullname string, pubkey, key []byte) (*User, error) {
	return nil, nil
}
