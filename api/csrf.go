package api

import (
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/keydotcat/server/util"
)

const CSRF_COOKIE_NAME = "kc4d018d7e07"

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
	if cookie, err := r.Cookie(CSRF_COOKIE_NAME); err == nil {
		var csrfToken string
		if err = c.sc.Decode(CSRF_COOKIE_NAME, cookie.Value, &csrfToken); err == nil {
			return csrfToken, false
		}
	}
	return c.generateNewToken(w), true
}

func (c csrf) generateNewToken(w http.ResponseWriter) string {
	csrfToken := util.GenerateRandomToken(8)
	if encoded, err := c.sc.Encode(CSRF_COOKIE_NAME, csrfToken); err == nil {
		cookie := &http.Cookie{
			Name:     CSRF_COOKIE_NAME,
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	} else {
		panic(err)
	}
	return csrfToken
}
