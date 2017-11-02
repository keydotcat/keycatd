package api

import (
	"net/http"

	"github.com/keydotcat/backend/util"
)

func (ah apiHandler) userRoot(w http.ResponseWriter, r *http.Request) {
	var head string
	var err error
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "":
		switch r.Method {
		case "GET":
			err = ah.userGetInfo(w, r)
		default:
			err = util.NewErrorFrom(ErrNotFound)
		}
	}
	if err != nil {
		httpErr(w, err)
	}
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
