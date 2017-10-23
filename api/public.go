package api

import (
	"encoding/json"
	"net/http"

	"github.com/keydotcat/backend/models"
)

type apiRegisterRequest struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Fullname       string `json:"fullname"`
	PublicKey      string `json:"public_key"`
	Password       string `json:"password"`
	Key            string `json:"key"`
	VaultPublicKey string `json:"vault_public_key"`
	VaultKey       string `json:"vault_key"`
}

func apiRegister(w http.ResponseWriter, r *http.Request) {
	apr := &apiRegisterRequest
	if err := jsonDecode(w, r, 1024*5, apr); err != nil {
		jsonErr(w)
		return
	}
	_, t, err := models.NewUser(
		r.Context(),
		apr.Id,
		apr.Fullname,
		apr.Email,
		apr.Password,
		apr.PublicKey,
		apr.Key,
		models.VaultKeyPair{
			PublicKey: apr.VaultPublicKey,
			Keys:      map[string][]byte{apr.Id: apr.VaultKey},
		},
	)
	if err != nil {
		httpErr(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(t)
}
