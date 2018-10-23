package api

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/keydotcat/keycatd/models"
	"github.com/keydotcat/keycatd/util"
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
	if err == nil {
		return false
	}
	buf := util.BufPool.Get()
	defer util.BufPool.Put(buf)
	json.NewEncoder(buf).Encode(err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.String())))
	if util.CheckErr(err, ErrNotFound) || util.CheckErr(err, models.ErrDoesntExist) {
		w.WriteHeader(http.StatusNotFound)
	} else if util.CheckErr(err, models.ErrUnauthorized) {
		w.WriteHeader(http.StatusUnauthorized)
	} else if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	buf.WriteTo(w)
	return true
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
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, max)).Decode(obj); err != nil {
		log.Printf("[ERROR] Could not parse json: %s", err)
		return util.NewErrorf("Could not parse request. Probably malformed")
	}
	return nil
}
