package httpx

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

// Router is an httpx.Handler router.
type Router struct {
	// NotFoundHandler is a Handler that will be called when a route is not
	// found.
	NotFoundHandler Handler

	// This router is ultimately backed by a gorilla mux router.
	mux *mux.Router
}

// NewRouter returns a new Router instance.
func NewRouter() *Router {
	return &Router{
		mux: mux.NewRouter(),
	}
}

// Handle registers a new route with a matcher for the URL path
func (r *Router) Handle(path string, h Handler) *Route {
	return &Route{r.mux.Handle(path, r.handler(h))}
}

// HandleFunc registers a new route with a matcher for the URL path
func (r *Router) HandleFunc(path string, f func(context.Context, http.ResponseWriter, *http.Request) error) *Route {
	return r.Handle(path, HandlerFunc(f))
}

// Header adds a route that will be used if the header value matches.
func (r *Router) Headers(pairs ...string) *Route {
	return &Route{r.mux.Headers(pairs...)}
}

// Match adds a route that will be matched if f returns true.
func (r *Router) Match(f func(*http.Request) bool, h Handler) {
	matcher := func(r *http.Request, rm *mux.RouteMatch) bool {
		return f(r)
	}

	r.mux.MatcherFunc(matcher).Handler(r.handler(h))
}

// mux.Handler expects an http.Handler. We wrap the Hander in a handler,
// which satisfies the http.Handler interface. When this route is
// eventually used, it's type asserted back to a Handler.
func (r *Router) handler(h Handler) http.Handler {
	return &handler{h}
}

// Handler returns a Handler that can be used to serve the request. Most of this
// is pulled from http://goo.gl/tyxad8.
func (r *Router) Handler(req *http.Request) (h Handler, vars map[string]string) {
	var match mux.RouteMatch

	if r.mux.Match(req, &match) {
		h = match.Handler.(Handler)
		vars = match.Vars
		return
	}

	if r.NotFoundHandler == nil {
		h = HandlerFunc(NotFound)
		return
	}

	h = r.NotFoundHandler
	return
}

// ServeHTTPContext implements the Handler interface.
func (r *Router) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
	h, vars := r.Handler(req)
	return h.ServeHTTPContext(WithVars(ctx, vars), w, req)
}

// Vars extracts the route vars from a context.Context.
func Vars(ctx context.Context) map[string]string {
	vars, ok := ctx.Value(varsKey).(map[string]string)
	if !ok {
		return map[string]string{}
	}

	return vars
}

// WithVars adds the vars to the context.Context.
func WithVars(ctx context.Context, vars map[string]string) context.Context {
	return context.WithValue(ctx, varsKey, vars)
}

// Route wraps a mux.Route.
type Route struct {
	route *mux.Route
}

// Methods adds a matcher for HTTP methods.
// It accepts a sequence of one or more methods to be matched, e.g.:
// "GET", "POST", "PUT".
func (r *Route) Methods(methods ...string) *Route {
	return &Route{r.route.Methods(methods...)}
}

// HandlerFunc sets the httpx.Handler for this route.
func (r *Route) HandlerFunc(f func(context.Context, http.ResponseWriter, *http.Request) error) *Route {
	return r.Handler(HandlerFunc(f))
}

// Handler sets the httpx.Handler for this route.
func (r *Route) Handler(h Handler) *Route {
	return &Route{r.route.Handler(r.handler(h))}
}

// mux.Handler expects an http.Handler. We wrap the Hander in a handler,
// which satisfies the http.Handler interface. When this route is
// eventually used, it's type asserted back to a Handler.
func (r *Route) handler(h Handler) http.Handler {
	return &handler{h}
}

// handler adapts a Handler to an http.Handler.
type handler struct {
	Handler
}

// ServeHTTP implements the http.Handler interface. This method is never
// actually called by this package, it's only used as a means to pass a Handler
// in and out of mux.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic(fmt.Sprintf("httpx: ServeHTTP called on %v", h))
}

// NotFound is a HandlerFunc that just delegates off to http.NotFound.
func NotFound(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	http.NotFound(w, r)
	return nil
}
