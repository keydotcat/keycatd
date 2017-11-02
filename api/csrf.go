package api

import (
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/keydotcat/backend/util"
)

const CSRF_COOKIE_NAME = "4d018d7e07"

type csrf struct {
	sc *securecookie.SecureCookie
}

func newCsrf(hKey, bKey []byte) csrf {
	return csrf{securecookie.New(hKey, bKey)}
}

func (c csrf) checkToken(w http.ResponseWriter, r *http.Request) (string, bool) {
	token, generated := c.getToken(w, r)
	if generated {
		return token, true
	}
	val, ok := r.Header["X-Csrf-Token"]
	if !ok {
		return token, false
	}
	if len(val) == 0 {
		return token, false
	}
	return token, token == val[0]
}

func (c csrf) getToken(w http.ResponseWriter, r *http.Request) (string, bool) {
	csrfToken := ""
	if cookie, err := r.Cookie(CSRF_COOKIE_NAME); err == nil {
		if err = c.sc.Decode(CSRF_COOKIE_NAME, cookie.Value, &csrfToken); err == nil {
			return csrfToken, false
		}
	}
	csrfToken = util.GenerateRandomToken(32)
	if encoded, err := c.sc.Encode(CSRF_COOKIE_NAME, csrfToken); err == nil {
		cookie := &http.Cookie{
			Name:  CSRF_COOKIE_NAME,
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	} else {
		panic(err)
	}
	return csrfToken, true
}
