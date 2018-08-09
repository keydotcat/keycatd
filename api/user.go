package api

import (
	"net/http"

	"github.com/keydotcat/backend/util"
)

func (ah apiHandler) userRoot(w http.ResponseWriter, r *http.Request) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		switch r.Method {
		case "GET":
			return ah.userGetInfo(w, r)
		case "PUT", "PATCH":
			return ah.userUpdate(w, r)
		}
	}
	return util.NewErrorFrom(ErrNotFound)
}

// GET /user
func (ah apiHandler) userGetInfo(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	currentUser := ctxGetUser(ctx)
	uf, err := currentUser.GetUserFull(ctx)
	if err != nil {
		return err
	}
	return jsonResponse(w, uf)
}

type userUpdateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	KeyPack  []byte `json:"user_keys"`
}

func (ah apiHandler) userUpdate(w http.ResponseWriter, r *http.Request) error {
	uur := &userUpdateRequest{}
	if err := jsonDecode(w, r, 8192, uur); err != nil {
		return err
	}
	ctx := r.Context()
	u := ctxGetUser(ctx)
	if len(uur.Email) > 3 {
		t, err := u.ChangeEmail(ctx, uur.Email)
		if err != nil {
			return err
		}
		if err := ah.mail.sendConfirmationMail(u, t, r.Header.Get("X-Locale")); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
	if len(uur.Password) > 0 {
		err := u.ChangePassword(ctx, uur.Password, uur.KeyPack)
		if err != nil {
			return err
		}
		w.WriteHeader(http.StatusOK)
		return nil
	}
	return util.NewErrorFrom(ErrNotFound)
}
