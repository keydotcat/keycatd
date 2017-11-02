package api

import (
	"net/http"

	"github.com/keydotcat/backend/managers"
	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

func (ah apiHandler) sessionRoot(w http.ResponseWriter, r *http.Request) {
	var head string
	var err error
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "":
		//TODO: list all sessions
		err = util.NewErrorFrom(ErrNotFound)
	default:
		//Actions over a token
		switch r.Method {
		case "GET":
			err = ah.sessionGetToken(w, r, head)
		default:
			err = util.NewErrorFrom(ErrNotFound)
		}
	}
	if err != nil {
		httpErr(w, err)
	}
}

type sessionResponse struct {
	managers.Session
	Csrf       string `json:"csrf,omitempty"`
	StoreToken string `json:"store_token,omitempty"`
}

// GET /session/:token
func (ah apiHandler) sessionGetToken(w http.ResponseWriter, r *http.Request, tid string) error {
	currentSession := ctxGetSession(r.Context())
	if currentSession.Id == tid {
		return jsonResponse(w, sessionResponse{currentSession, ctxGetCsrf(r.Context()), currentSession.StoreToken})
	}
	currentUser := ctxGetUser(r.Context())
	s, err := ah.sm.GetSession(tid)
	if err != nil {
		return util.NewErrorFrom(models.ErrDoesntExist)
	}
	if s.UserId != currentUser.Id {
		return util.NewErrorFrom(models.ErrDoesntExist)
	}
	return jsonResponse(w, s)
}
