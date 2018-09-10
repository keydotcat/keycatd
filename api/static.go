package api

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/keydotcat/server/static"
)

//HACK: html/template removes the html comments and since we only require the csrf.. :P

type StaticHandler struct {
	Dir         string
	IndexFile   string
	cacheStatic bool
}

func NewStaticHandler() *StaticHandler {
	return &StaticHandler{
		Dir:       "web",
		IndexFile: "index.html",
	}
}

func (s *StaticHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.NotFound(rw, r)
		return
	}
	file := r.URL.Path
	if strings.Index(file, "/api/") == 0 || strings.Index(file, "/oa/") == 0 {
		http.NotFound(rw, r)
		return
	}
	file = strings.TrimLeft(file, "/")
	if len(file) == 0 || file == s.IndexFile || len(filepath.Ext(file)) == 0 {
		file = s.IndexFile
	}
	filePath := fmt.Sprintf("%s/%s", s.Dir, file)
	finfo, err := static.AssetInfo(filePath)
	if err != nil {
		http.NotFound(rw, r)
		return
	}
	data, err := static.Asset(filePath)
	if err != nil {
		http.NotFound(rw, r)
		return
	}
	if s.cacheStatic {
		rw.Header().Add("Cache-Control", "public, max-age=31536000")
		rw.Header().Add("Expires", time.Now().Add(30*24*time.Hour).Format(time.RFC1123))
	}
	http.ServeContent(rw, r, file, finfo.ModTime(), bytes.NewReader(data))
}
