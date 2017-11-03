package api

import (
	"net/http"

	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

func (ah apiHandler) teamRoot(w http.ResponseWriter, r *http.Request) {
	var head string
	var err error
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "":
		switch r.Method {
		case "GET":
			err = ah.teamGetInfo(w, r)
		case "POST":
			err = ah.teamCreate(w, r)
		default:
			err = util.NewErrorFrom(ErrNotFound)
		}
	}
	if err != nil {
		httpErr(w, err)
	}
}

type teamGetInfoResponse struct {
	Teams []*models.Team `json:"teams"`
}

// GET /team
func (ah apiHandler) teamGetInfo(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	currentUser := ctxGetUser(ctx)
	teams, err := currentUser.GetTeams(ctx)
	if err != nil {
		return err
	}
	return jsonResponse(w, teamGetInfoResponse{teams})
}

type teamCreateRequest struct {
	Name      string              `json:"name"`
	VaultKeys models.VaultKeyPair `json:"vault_keys"`
}

// POST /team
func (ah apiHandler) teamCreate(w http.ResponseWriter, r *http.Request) error {
	tcr := &teamCreateRequest{}
	if err := jsonDecode(w, r, 4096, tcr); err != nil {
		return err
	}
	ctx := r.Context()
	currentUser := ctxGetUser(ctx)
	team, err := currentUser.CreateTeam(ctx, tcr.Name, tcr.VaultKeys)
	if err != nil {
		return err
	}
	tf, err := team.GetTeamFull(ctx, currentUser)
	if err != nil {
		return err
	}
	return jsonResponse(w, tf)
}
