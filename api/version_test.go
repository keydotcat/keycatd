package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/keydotcat/server/util"
)

func TestGetFullVersion(t *testing.T) {
	r, err := http.Get(srv.URL + "/version")
	CheckErrorAndResponse(t, r, err, 200)
	sga := &versionSendFullResponse{}
	if err := json.NewDecoder(r.Body).Decode(sga); err != nil {
		t.Fatal(err)
	}
	if sga.Server != util.GetServerVersion() {
		t.Errorf("Mismatch in the server version: %s vs %s", util.GetServerVersion(), sga.Server)
	}
	if sga.Web != util.GetWebVersion() {
		t.Errorf("Mismatch in the web version: %s vs %s", util.GetWebVersion(), sga.Web)
	}
}
