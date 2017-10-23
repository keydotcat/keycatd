package api

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"
)

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func jsonErr(w ResponseWriter) {
	http.Error(w, "Could not decode JSON data", http.StatusBadRequest)
}

func httpErr(w ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func jsonDecode(w http.ResponseWrite, r *http.Request, max int64, obj interface{}) error {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, max)).Decode(obj); err != nil {
		return err
	}
}
