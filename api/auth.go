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
		if csrfToken, valid := ah.csrf.checkToken(w, r); !valid {
			http.Error(w, "Invalid CSRF token", http.StatusUnauthorized)
			return nil
		} else {
			r = r.WithContext(ctxAddCsrf(r.Context(), csrfToken))
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
	return r.WithContext(ctxAddUser(ctxAddSession(r.Context(), s), u))
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

func (ah apiHandler) authRoot(w http.ResponseWriter, r *http.Request) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "register":
		return ah.authRegister(w, r)
	case "confirm_email":
		return ah.authConfirmEmail(w, r)
	case "request_confirmation_token":
		return ah.authRequestConfirmationToken(w, r)
	case "login":
		return ah.authLogin(w, r)
	}
	return util.NewErrorFrom(ErrNotFound)
}

// /auth/register
func (ah apiHandler) authRegister(w http.ResponseWriter, r *http.Request) error {
	apr := &authRegisterRequest{}
	if err := jsonDecode(w, r, 1024*5, apr); err != nil {
		return err
	}
	u, t, err := models.NewUser(
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
	if err := ah.mail.sendConfirmationMail(u, t, r.Header.Get("X-Locale")); err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	return nil
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
		return util.NewErrorFrom(models.ErrDoesntExist)
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
	if err := ah.mail.sendConfirmationMail(u, t, r.Header.Get("X-Locale")); err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

type authLoginResponse struct {
	Username   string `json:"user_id"`
	Token      string `json:"session_token"`
	StoreToken string `json:"store_token"`
	PublicKeys []byte `json:"public_key"`
	SecretKeys []byte `json:"secret_key"`
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
	if !u.ConfirmedAt.Valid {
		return util.NewErrorFrom(models.ErrUnauthorized)
	}
	if err := u.CheckPassword(aer.Password); err != nil {
		return util.NewErrorFrom(models.ErrUnauthorized)
	}
	s, err := ah.sm.NewSession(u.Id, r.UserAgent(), aer.RequireCSRF)
	if err != nil {
		panic(err)
	}
	return jsonResponse(w, authLoginResponse{
		u.Id,
		s.Id,
		s.StoreToken,
		u.PublicKey,
		u.Key,
	})
}
