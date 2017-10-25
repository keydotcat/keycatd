package api

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func jsonErr(w http.ResponseWriter) {
	http.Error(w, "Could not decode JSON data", http.StatusBadRequest)
}

func httpErr(w http.ResponseWriter, err error) bool {
	if util.CheckErr(err, models.ErrDoesntExist) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return true
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return true
	}
	return false
}

func jsonResponse(w http.ResponseWriter, obj interface{}) error {
	b := util.BufPool.Get()
	defer util.BufPool.Put(b)
	if err := json.NewEncoder(b).Encode(obj); err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(b.Bytes())))
	w.WriteHeader(http.StatusOK)
	b.WriteTo(w)
	return nil
}

func jsonDecode(w http.ResponseWriter, r *http.Request, max int64, obj interface{}) error {
	return json.NewDecoder(http.MaxBytesReader(w, r.Body, max)).Decode(obj)
}
