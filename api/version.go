package api

import (
	"net/http"

	"github.com/keydotcat/server/util"
)

func (ah apiHandler) versionRoot(w http.ResponseWriter, r *http.Request) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		return ah.versionSendFull(w, r)
	}
	return util.NewErrorFrom(ErrNotFound)
}

type versionSendFullResponse struct {
	Name   string `json:"name"`
	Server string `json:"server"`
	Web    string `json:"web"`
}

// /version
func (ah apiHandler) versionSendFull(w http.ResponseWriter, r *http.Request) error {
	return jsonResponse(w, versionSendFullResponse{Name: "Key cat", Server: util.GetServerVersion(), Web: util.GetWebVersion()})
}
