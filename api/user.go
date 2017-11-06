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
