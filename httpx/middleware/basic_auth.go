package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

type BasicAuther struct {
	User, Pass string
	Realm      string

	// The handler that will be called if the request is authorized.
	Handler http.Handler

	// The handler that will be called if the request is not authorized. The
	// zero value is DefaultUnauthorizedHandler
	UnauthorizedHandler http.Handler
}

func (a *BasicAuther) authenticated(r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Basic" {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}
	return a.User == pair[0] && a.Pass == pair[1]
}

func DefaultUnauthorizedHandler(realm string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%q`, realm))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

func (a *BasicAuther) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.authenticated(r) {
		a.Handler.ServeHTTP(w, r)
	} else {
		u := a.UnauthorizedHandler
		if u == nil {
			u = DefaultUnauthorizedHandler(a.Realm)
		}
		u.ServeHTTP(w, r)
	}
}

func BasicAuth(h http.Handler, user, pass, realm string) *BasicAuther {
	return &BasicAuther{
		User:    user,
		Pass:    pass,
		Realm:   realm,
		Handler: h,
	}
}
