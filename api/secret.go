package api

import (
	"fmt"
	"net/http"

	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

// /team/:tid/secret
func (ah apiHandler) teamSecretRoot(w http.ResponseWriter, r *http.Request, t *models.Team) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		switch r.Method {
		case "GET":
			return ah.teamSecretGetAll(w, r, t)
		}
	}
	return util.NewErrorFrom(ErrNotFound)
}

type teamSecretGetAllResponse struct {
	Secrets []*models.Secret `json:"secrets"`
}

// GET /team/:tid/secret
func (ah apiHandler) teamSecretGetAll(w http.ResponseWriter, r *http.Request, t *models.Team) error {
	ctx := r.Context()
	u := ctxGetUser(ctx)
	s, err := t.GetSecretsForUser(ctx, u)
	if err != nil {
		return err
	}
	return jsonResponse(w, teamSecretGetAllResponse{s})
}

// /team/:tid/vault/:vid/secret
func (ah apiHandler) validVaultSecretRoot(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		switch r.Method {
		case "GET":
			//TODO: Get all secrets for vault
		case "POST":
			return ah.vaultCreateSecret(w, r, t, v)
		}
	} else {
		fmt.Println("Got to root", head)
		switch r.Method {
		case "DELETE":
			return ah.vaultDeleteSecret(w, r, t, v, head)
		}
	}
	return util.NewErrorFrom(ErrNotFound)
}

type vaultCreateSecretRequest struct {
	Data []byte `json:"data"`
}

func (ah apiHandler) vaultCreateSecret(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault) error {
	ctx := r.Context()
	vscr := &vaultCreateSecretRequest{}
	if err := jsonDecode(w, r, 16*1024, vscr); err != nil {
		return err
	}
	s := &models.Secret{Data: vscr.Data}
	if err := v.AddSecret(ctx, s); err != nil {
		return err
	}
	return jsonResponse(w, s)
}

func (ah apiHandler) vaultDeleteSecret(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault, sid string) error {
	ctx := r.Context()
	fmt.Println("Got to func", sid)
	if err := v.DeleteSecret(ctx, sid); err != nil {
		return err
	}
	return jsonResponse(w, v)
}
