package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/keydotcat/backend/models"
)

func MainHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, "-", r.RequestURI)
		apiRootHandler(w, r.WithContext(models.AddDBToContext(r.Context(), db)))
	})
}

func apiRootHandler(w http.ResponseWriter, r *http.Request) {
	head := ""
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "auth":
		apiAuth(w, r)
	default:
		http.NotFound(w, r)
	}
}
