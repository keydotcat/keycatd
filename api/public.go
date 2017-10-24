package api

import (
	"net/http"

	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

type apiAuthRegisterRequest struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Fullname       string `json:"fullname"`
	Password       string `json:"password"`
	PublicKey      []byte `json:"public_key"`
	Key            []byte `json:"key"`
	VaultPublicKey []byte `json:"vault_public_key"`
	VaultKey       []byte `json:"vault_key"`
}

func apiAuth(w http.ResponseWriter, r *http.Request) {
	var head string
	var err error
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "register":
		err = apiAuthRegister(w, r)
	case "confirm_email":
		err = apiAuthConfirmEmail(w, r)
	case "request_confirmation_token":
		err = apiAuthRequestConfirmationToken(w, r)
	default:
		err = models.ErrDoesntExist
	}
	if err != nil {
		httpErr(w, err)
	}
}

// /auth/register
func apiAuthRegister(w http.ResponseWriter, r *http.Request) error {
	apr := &apiAuthRegisterRequest{}
	if err := jsonDecode(w, r, 1024*5, apr); err != nil {
		return err
	}
	_, t, err := models.NewUser(
		r.Context(),
		apr.Id,
		apr.Fullname,
		apr.Email,
		apr.Password,
		apr.PublicKey,
		apr.Key,
		models.VaultKeyPair{
			PublicKey: apr.VaultPublicKey,
			Keys:      map[string][]byte{apr.Id: apr.VaultKey},
		},
	)
	if err != nil {
		return err
	}
	//TODO: Send email
	return jsonResponse(w, t)
}

// /auth/confirm_email/:token
func apiAuthConfirmEmail(w http.ResponseWriter, r *http.Request) error {
	token, _ := shiftPath(r.URL.Path)
	if len(token) == 0 {
		return util.NewErrorFrom(models.ErrDoesntExist)
	}
	tok, err := models.FindToken(r.Context(), token)
	if err != nil {
		return err
	}
	u, err := tok.ConfirmEmail(r.Context())
	if err != nil {
		return err
	}
	return jsonResponse(w, u)
}

type authEmailRequest struct {
	Email string `json:"email"`
}

// /auth/request_confirmation_token
func apiAuthRequestConfirmationToken(w http.ResponseWriter, r *http.Request) error {
	aer := &authEmailRequest{}
	if err := jsonDecode(w, r, 1024, aer); err != nil {
		return err
	}
	u, err := models.FindUserByEmail(r.Context(), aer.Email)
	if err != nil {
		return err
	}
	t, err := u.GetVerificationToken(r.Context())
	if err != nil {
		return err
	}
	//TODO: Send email
	return jsonResponse(w, t)
}
