package api

import (
	"net/http"

	"github.com/keydotcat/keycatd/managers"
	"github.com/keydotcat/keycatd/models"
	"github.com/keydotcat/keycatd/util"
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

type teamSecretListWrap struct {
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
	return jsonResponse(w, teamSecretListWrap{s})
}

// /team/:tid/vault/:vid/secret
func (ah apiHandler) validVaultSecretRoot(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		switch r.Method {
		case "GET":
			return ah.vaultGetSecrets(w, r, t, v)
		case "POST":
			return ah.vaultCreateSecret(w, r, t, v)
		}
	} else {
		switch r.Method {
		case "DELETE":
			return ah.vaultDeleteSecret(w, r, t, v, head)
		case "PATCH", "PUT":
			return ah.vaultUpdateSecret(w, r, t, v, head)
		}
	}
	return util.NewErrorFrom(ErrNotFound)
}

func (ah apiHandler) vaultGetSecrets(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault) error {
	ctx := r.Context()
	secrets, err := v.GetSecrets(ctx)
	if err != nil {
		return err
	}
	return jsonResponse(w, teamSecretListWrap{secrets})

}

type vaultCreateSecretRequest struct {
	Team  string `json:"team"`
	Vault string `json:"vault"`
	Data  []byte `json:"data"`
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
	ah.bcast.Send(v.Team, v.Id, managers.BCAST_ACTION_SECRET_NEW, s)
	return jsonResponse(w, s)
}

func (ah apiHandler) vaultDeleteSecret(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault, sid string) error {
	ctx := r.Context()
	if err := v.DeleteSecret(ctx, sid); err != nil {
		return err
	}
	ah.bcast.Send(v.Team, v.Id, managers.BCAST_ACTION_SECRET_REMOVE, &models.Secret{Id: sid})
	return jsonResponse(w, v)
}

func (ah apiHandler) vaultUpdateSecret(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault, sid string) error {
	ctx := r.Context()
	vscr := &vaultCreateSecretRequest{}
	if err := jsonDecode(w, r, 16*1024, vscr); err != nil {
		return err
	}
	//Modify secret
	if len(vscr.Data) > 0 {
		s := &models.Secret{Id: sid, Data: vscr.Data}
		if err := v.UpdateSecret(ctx, s); err != nil {
			return err
		}
		if len(vscr.Vault) == 0 {
			ah.bcast.Send(v.Team, v.Id, managers.BCAST_ACTION_SECRET_CHANGE, s)
			return jsonResponse(w, s)
		}
	}
	//Move it to a different team/vault
	u := ctxGetUser(r.Context())
	var targetTeam = t
	if len(vscr.Team) != 0 {
		var err error
		targetTeam, err = u.GetTeam(r.Context(), vscr.Team)
		if err != nil {
			return err
		}
	}
	targetVault, err := targetTeam.GetVaultForUser(r.Context(), vscr.Vault, u)
	if err != nil {
		return err
	}
	s := &models.Secret{Team: t.Id, Vault: v.Id, Id: sid}
	if err := s.MoveToTeamVault(ctx, targetTeam.Id, targetVault.Id); err != nil {
		return err
	}
	ah.bcast.Send(v.Team, v.Id, managers.BCAST_ACTION_SECRET_REMOVE, s)
	ah.bcast.Send(targetTeam.Id, targetVault.Id, managers.BCAST_ACTION_SECRET_NEW, s)
	return jsonResponse(w, s)

}

// /team/:tid/vault/:vid/secrets
func (ah apiHandler) validVaultSecretsRoot(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		switch r.Method {
		case "POST":
			return ah.vaultCreateSecretList(w, r, t, v)
		}
	}
	return util.NewErrorFrom(ErrNotFound)
}

func (ah apiHandler) vaultCreateSecretList(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault) error {
	ctx := r.Context()
	vl := &teamSecretListWrap{}
	if err := jsonDecode(w, r, 1024*1024, &vl); err != nil {
		return err
	}
	sl := make([]*models.Secret, len(vl.Secrets))
	for i, vc := range vl.Secrets {
		sl[i] = &models.Secret{Data: vc.Data}
	}
	if err := v.AddSecretList(ctx, sl); err != nil {
		return err
	}
	for _, s := range sl {
		ah.bcast.Send(v.Team, v.Id, managers.BCAST_ACTION_SECRET_NEW, s)
	}
	return jsonResponse(w, teamSecretListWrap{sl})
}
