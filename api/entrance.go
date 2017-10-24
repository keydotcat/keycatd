package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/keydotcat/backend/models"
)

type apiHandler struct {
	db   *sql.DB
	sm   SessionManager
	csrf CSRF
}

func NewAPIHander(db *sql.DB, sm SessionManager, c CSRF) http.Handler {
	ah := apiHandler{db, sm, c}
	return ah
}

func (ah apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, "-", r.RequestURI)
	r = r.WithContext(models.AddDBToContext(r.Context(), ah.db))
	head := ""
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "auth":
		ah.authRoot(w, r)
	default:
		http.NotFound(w, r)
	}
}
