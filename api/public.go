package api

import (
	"net/http"

	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

type apiRegisterRequest struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Fullname       string `json:"fullname"`
	PublicKey      string `json:"public_key"`
	Password       string `json:"password"`
	Key            string `json:"key"`
	VaultPublicKey string `json:"vault_public_key"`
	VaultKey       string `json:"vault_key"`
}

func apiAuth(w htt.ResponseWriter, r *http.Request) {
	head, req.URL.Path = ShiftPath(req.URL.Path)
	switch head {
	case "register":
		apiAuthRegister(w, r)
	case "confirm_email":
		apiAuthConfirmEmail(w, r)
	case "request_confirmation_token":
		apiAuthRequestConfirmationToken(w, r)
	default:
		http.NotFound(w, r)
	}

}

// /auth/register
func apiAuthRegister(w http.ResponseWriter, r *http.Request) error {
	apr := &apiRegisterRequest
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
	token, _ := splitPath(r.URL.Path)
	if len(token) == 0 {
		return util.NewErrorFrom(models.ErrDoesntExist)
	}
	tok, err := models.FindToken(r.GetContext(), token)
	if err != nil {
		return err
	}
	u, err := tok.ConfirmEmail(r.GetContext())
	if err != nil {
		return err
	}
	return jsonResponse(w, t)
}

type authEmailRequest struct {
	Email string `json:"email"`
}

// /auth/request_confirmation_token
func apiAuthConfirmEmail(w http.ResponseWriter, r *http.Request) {
	aer := &apiEmailRequest{}
	if err := jsonDecode(w, r, 1024, aer); err != nil {
		return err
	}
	u, err := models.FindUserByEmail(r.GetContext(), aer.Email)
	if err != nil {
		return err
	}
	t, err := u.GetVerificationToken(r.GetContext())
	if err != nil {
		return err
	}
	//TODO: Send email
	return jsonResponse(w, t)
}
