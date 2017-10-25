package api

import (
	"net/http"

	"github.com/keydotcat/backend/models"
)

func (ah apiHandler) teamRoot(w http.ResponseWriter, r *http.Request) {
	if r = ah.authorizeRequest(w, r); r == nil {
		return
	}
	var head string
	var err error
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "":
		switch r.Method {
		case "GET":
			err = ah.teamList(w, r)
		case "POST":
			err = ah.teamCreate(w, r)
		default:
			err = models.ErrDoesntExist
		}
	default:
		err = models.ErrDoesntExist
	}
	if err != nil {
		httpErr(w, err)
	}
}

type teamListResponse struct {
	Teams []*models.Team `json:"teams"`
}

// GET /team
func (ah apiHandler) teamList(w http.ResponseWriter, r *http.Request) error {
	u := ctxGetUser(r.Context())
	teams, err := u.GetTeams(r.Context())
	if err != nil {
		return err
	}
	return jsonResponse(w, teamListResponse{teams})
}

type teamCreateRequest struct {
	Name      string `json:"name"`
	PublicKey []byte `json:"public_key"`
	Key       []byte `json:"key"`
}

// POST /team
func (ah apiHandler) teamCreate(w http.ResponseWriter, r *http.Request) error {
	u := ctxGetUser(r.Context())
	teams, err := u.GetTeams(r.Context())
	if err != nil {
		return err
	}
	return jsonResponse(w, teamListResponse{teams})
}
