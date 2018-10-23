package api

import (
	"net/http"

	"github.com/keydotcat/keycatd/models"
	"github.com/keydotcat/keycatd/util"
)

// /team/:tid/vault
func (ah apiHandler) vaultRoot(w http.ResponseWriter, r *http.Request, t *models.Team) error {
	var vid string
	vid, r.URL.Path = shiftPath(r.URL.Path)
	if len(vid) == 0 {
		switch r.Method {
		case "GET":
			return ah.vaultList(w, r, t)
		case "POST":
			return ah.vaultCreate(w, r, t)
		}
	} else {
		u := ctxGetUser(r.Context())
		v, err := t.GetVaultForUser(r.Context(), vid, u)
		if err != nil {
			return err
		}
		return ah.validVaultRoot(w, r, t, v)
	}
	return util.NewErrorFrom(ErrNotFound)
}

type vaultListResponse struct {
	Vaults []*models.VaultFull `json:"vaults"`
}

func (ah apiHandler) vaultList(w http.ResponseWriter, r *http.Request, t *models.Team) error {
	ctx := r.Context()
	u := ctxGetUser(ctx)
	vs, err := t.GetVaultsFullForUser(ctx, u)
	if err != nil {
		return err
	}
	return jsonResponse(w, vaultListResponse{vs})
}

type vaultCreateRequest struct {
	Name string              `json:"name"`
	Keys models.VaultKeyPair `json:"vault_keys"`
}

func (ah apiHandler) vaultCreate(w http.ResponseWriter, r *http.Request, t *models.Team) error {
	vcr := &vaultCreateRequest{}
	if err := jsonDecode(w, r, 81920, vcr); err != nil {
		return err
	}
	ctx := r.Context()
	u := ctxGetUser(ctx)
	v, err := t.CreateVault(ctx, u, vcr.Name, vcr.Keys)
	if err != nil {
		return err
	}
	vf, err := v.GetVaultFullForUser(ctx, u)
	if err != nil {
		return err
	}
	return jsonResponse(w, vf)
}

// /team/:tid/vault/:vid
func (ah apiHandler) validVaultRoot(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		//Not yet here
	} else {
		switch head {
		case "user":
			return ah.validVaultUserRoot(w, r, t, v)
		case "secret":
			return ah.validVaultSecretRoot(w, r, t, v)
		case "secrets":
			return ah.validVaultSecretsRoot(w, r, t, v)
		}
	}
	return util.NewErrorFrom(ErrNotFound)
}

// /team/:tid/vault/:vid/user
func (ah apiHandler) validVaultUserRoot(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault) error {
	var uid string
	uid, r.URL.Path = shiftPath(r.URL.Path)
	if len(uid) == 0 {
		switch r.Method {
		case "POST":
			return ah.vaultAddUser(w, r, t, v)
		}
	} else {
		switch r.Method {
		case "DELETE":
			return ah.vaultRemoveUser(w, r, t, v, uid)
		}
	}
	return util.NewErrorFrom(ErrNotFound)
}

// POST /team/:tid/vault/:vid/user
func (ah apiHandler) vaultAddUser(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault) error {
	var keys map[string][]byte
	if err := jsonDecode(w, r, 2048, &keys); err != nil {
		return err
	}
	ctx := r.Context()
	u := ctxGetUser(ctx)
	isAdmin, err := t.CheckAdmin(ctx, u)
	if err != nil {
		return err
	}
	if !isAdmin {
		return util.NewErrorFrom(models.ErrUnauthorized)
	}
	if err := v.AddUsers(ctx, keys); err != nil {
		return err
	}
	vf, err := v.GetVaultFullForUser(ctx, u)
	if err != nil {
		return err
	}
	return jsonResponse(w, vf)
}

// DELETE /team/:tid/vault/:vid/user/:uid
func (ah apiHandler) vaultRemoveUser(w http.ResponseWriter, r *http.Request, t *models.Team, v *models.Vault, uid string) error {
	ctx := r.Context()
	u := ctxGetUser(ctx)
	isAdmin, err := t.CheckAdmin(ctx, u)
	if err != nil {
		return err
	}
	if !isAdmin {
		return util.NewErrorFrom(models.ErrUnauthorized)
	}
	if err := v.RemoveUser(ctx, uid); err != nil {
		return err
	}
	vf, err := v.GetVaultFullForUser(ctx, u)
	if err != nil {
		return err
	}
	return jsonResponse(w, vf)
}
