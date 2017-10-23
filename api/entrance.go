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
		apiRootHandler(w, r.WithContext(models.AddDBToContext(ctx, db)))
	})
}

func apiRootHandler(w ResponseWriter, r *Request) {
	head, req.URL.Path = ShiftPath(req.URL.Path)
	switch head {
	case "register":
		apiRegister(w, r)
	default:
		http.NotFound(w, r)
	}
}
