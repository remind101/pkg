package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/remind101/pkg/httpx"
	"golang.org/x/net/context"
)

type BasicAuther struct {
	User, Pass string
	Realm      string

	// The handler that will be called if the request is authorized.
	Handler httpx.Handler

	// The handler that will be called if the request is not authorized. The
	// zero value is DefaultUnauthorizedHandler
	UnauthorizedHandler httpx.Handler
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

func DefaultUnauthorizedHandler(realm string) httpx.HandlerFunc {
	return httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%q`, realm))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return nil
	})
}

func (a *BasicAuther) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if a.authenticated(r) {
		return a.Handler.ServeHTTPContext(ctx, w, r)
	} else {
		u := a.UnauthorizedHandler
		if u == nil {
			u = DefaultUnauthorizedHandler(a.Realm)
		}
		return u.ServeHTTPContext(ctx, w, r)
	}
}

func BasicAuth(h httpx.Handler, user, pass, realm string) *BasicAuther {
	return &BasicAuther{
		User:    user,
		Pass:    pass,
		Realm:   realm,
		Handler: h,
	}
}

func BasicAuthWithUserPass(user, pass string) *BasicAuther {
	return &BasicAuther{
		User: user,
		Pass: pass,
	}
}

func (a *BasicAuther) Authenticate(h httpx.Handler) httpx.Handler {
	a.Handler = h
	return httpx.HandlerFunc(a.ServeHTTPContext)
}
