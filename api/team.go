package api

import (
	"net/http"

	"github.com/keydotcat/server/models"
	"github.com/keydotcat/server/util"
)

func (ah apiHandler) teamRoot(w http.ResponseWriter, r *http.Request) error {
	var tid string
	tid, r.URL.Path = shiftPath(r.URL.Path)
	if len(tid) == 0 {
		switch r.Method {
		case "GET":
			return ah.teamGetAll(w, r)
		case "POST":
			return ah.teamCreate(w, r)
		default:
			return util.NewErrorFrom(ErrNotFound)
		}
	} else {
		u := ctxGetUser(r.Context())
		t, err := u.GetTeam(r.Context(), tid)
		if err != nil {
			return err
		}
		return ah.validTeamRoot(w, r, t)
	}
}

type teamGetAllResponse struct {
	Teams []*models.Team `json:"teams"`
}

// GET /team
func (ah apiHandler) teamGetAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	currentUser := ctxGetUser(ctx)
	teams, err := currentUser.GetTeams(ctx)
	if err != nil {
		return err
	}
	return jsonResponse(w, teamGetAllResponse{teams})
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

func (ah apiHandler) validTeamRoot(w http.ResponseWriter, r *http.Request, t *models.Team) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		switch r.Method {
		case "GET":
			return ah.teamGetInfo(w, r, t)
		default:
			return util.NewErrorFrom(ErrNotFound)
		}
	} else {
		switch head {
		case "user":
			return ah.validTeamUserRoot(w, r, t)
		case "vault":
			return ah.vaultRoot(w, r, t)
		case "secret":
			return ah.teamSecretRoot(w, r, t)
		}
	}
	return util.NewErrorFrom(ErrNotFound)
}

// GET /team/:tid
func (ah apiHandler) teamGetInfo(w http.ResponseWriter, r *http.Request, t *models.Team) error {
	ctx := r.Context()
	currentUser := ctxGetUser(ctx)
	tf, err := t.GetTeamFull(ctx, currentUser)
	if err != nil {
		return err
	}
	return jsonResponse(w, tf)
}

func (ah apiHandler) validTeamUserRoot(w http.ResponseWriter, r *http.Request, t *models.Team) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		switch r.Method {
		case "POST":
			return ah.teamInviteUser(w, r, t)
		}
	} else {
		switch r.Method {
		case "PATCH":
			return ah.teamModifyUser(w, r, t, head)
		}
	}
	return util.NewErrorFrom(ErrNotFound)
}

type teamInviteUserRequest struct {
	Invite string `json:"invite"`
}

// POST /team/:tid/user
func (ah apiHandler) teamInviteUser(w http.ResponseWriter, r *http.Request, t *models.Team) error {
	ctx := r.Context()
	u := ctxGetUser(ctx)
	tcr := &teamInviteUserRequest{}
	if err := jsonDecode(w, r, 1024, tcr); err != nil {
		return err
	}
	invite, err := t.AddOrInviteUserByEmail(ctx, u, tcr.Invite)
	if err != nil && !util.CheckErr(err, models.ErrAlreadyInvited) {
		return err
	}
	if invite != nil {
		if err := ah.mail.sendInvitationMail(t, u, invite, r.Header.Get("X-Locale")); err != nil {
			panic(err)
		}
	}
	tf, err := t.GetTeamFull(ctx, u)
	if err != nil {
		return err
	}
	return jsonResponse(w, tf)
}

type teamModifyUserRequest struct {
	Admin bool              `json:"admin"`
	Keys  map[string][]byte `json:"keys"`
}

type teamModifyUserResponse struct {
	Team  string                 `json:"team"`
	Users []*models.TeamUserFull `json:"users"`
}

// PATCH /team/:tid/user/:uid
func (ah apiHandler) teamModifyUser(w http.ResponseWriter, r *http.Request, t *models.Team, uid string) error {
	tiur := &teamModifyUserRequest{}
	if err := jsonDecode(w, r, 2048, tiur); err != nil {
		return err
	}
	ctx := r.Context()
	admin := ctxGetUser(ctx)
	u, err := models.FindUser(ctx, uid)
	if err != nil {
		return err
	}
	if tiur.Admin {
		err = t.PromoteUser(ctx, admin, u, models.VaultKeyPair{Keys: tiur.Keys})
	} else {
		err = t.DemoteUser(ctx, admin, u)
	}
	if err != nil {
		return err
	}
	tuf, err := t.GetUsersAfiliationFull(ctx)
	if err != nil {
		return err
	}
	return jsonResponse(w, teamModifyUserResponse{t.Id, tuf})
}
