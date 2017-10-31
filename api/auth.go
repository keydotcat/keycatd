package api

import (
	"net/http"
	"strings"

	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

func (ah apiHandler) authorizeRequest(w http.ResponseWriter, r *http.Request) *http.Request {
	authHdr := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHdr) < 2 || authHdr[0] != "Bearer" {
		http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		return nil
	}
	s, err := ah.sm.UpdateSession(authHdr[1], r.UserAgent())
	if err != nil {
		http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		return nil
	}
	if s.RequiresCSRF {
		if !ah.csrf.checkToken(w, r) {
			http.Error(w, "Invalid CSRF token", http.StatusUnauthorized)
			return nil
		}
	}
	u, err := models.FindUser(r.Context(), s.UserId)
	if util.CheckErr(err, models.ErrDoesntExist) {
		http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		ah.sm.DeleteAllSessions(u.Id)
		return nil
	} else if err != nil {
		panic(err)
	}
	return r.WithContext(ctxAddUser(r.Context(), u))
}

type authRegisterRequest struct {
	Username       string `json:"id"`
	Email          string `json:"email"`
	Fullname       string `json:"fullname"`
	Password       string `json:"password"`
	KeyPack        []byte `json:"user_keys"`
	VaultPublicKey []byte `json:"vault_public_keys"`
	VaultKey       []byte `json:"vault_keys"`
}

func (ah apiHandler) authRoot(w http.ResponseWriter, r *http.Request) {
	var head string
	var err error
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "register":
		err = ah.authRegister(w, r)
	case "confirm_email":
		err = ah.authConfirmEmail(w, r)
	case "request_confirmation_token":
		err = ah.authRequestConfirmationToken(w, r)
	case "login":
		err = ah.authLogin(w, r)
	default:
		err = util.NewErrorFrom(ErrNotFound)
	}
	if err != nil {
		httpErr(w, err)
	}
}

// /auth/register
func (ah apiHandler) authRegister(w http.ResponseWriter, r *http.Request) error {
	apr := &authRegisterRequest{}
	if err := jsonDecode(w, r, 1024*5, apr); err != nil {
		return err
	}
	_, t, err := models.NewUser(
		r.Context(),
		apr.Username,
		apr.Fullname,
		apr.Email,
		apr.Password,
		apr.KeyPack,
		models.VaultKeyPair{
			PublicKey: apr.VaultPublicKey,
			Keys:      map[string][]byte{apr.Username: apr.VaultKey},
		},
	)
	if err != nil {
		return err
	}
	//TODO: Send email
	return jsonResponse(w, t)
}

// /auth/confirm_email/:token
func (ah apiHandler) authConfirmEmail(w http.ResponseWriter, r *http.Request) error {
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

type authRequest struct {
	Id          string `json:"id"`
	Password    string `json:"password"`
	RequireCSRF bool   `json:"want_csrf"`
	Email       string `json:"email"`
}

// /auth/request_confirmation_token
func (ah apiHandler) authRequestConfirmationToken(w http.ResponseWriter, r *http.Request) error {
	aer := &authRequest{}
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

// /auth/login
func (ah apiHandler) authLogin(w http.ResponseWriter, r *http.Request) error {
	aer := &authRequest{}
	if err := jsonDecode(w, r, 1024, aer); err != nil {
		return err
	}
	u, err := models.FindUser(r.Context(), aer.Id)
	if util.CheckErr(err, models.ErrDoesntExist) {
		return util.NewErrorFrom(models.ErrUnauthorized)
	} else if err != nil {
		return err
	}
	if err := u.CheckPassword(aer.Password); err != nil {
		return util.NewErrorFrom(models.ErrUnauthorized)
	}
	s, err := ah.sm.NewSession(u.Id, r.UserAgent(), aer.RequireCSRF)
	if err != nil {
		panic(err)
	}
	return jsonResponse(w, s)
}
