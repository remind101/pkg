package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/remind101/pkg/httpx"
)

type opentracingTest struct {
	// A function that adds Handlers to the router.
	routes func(*httpx.Router)

	// An http.Request to test.
	req *http.Request

	expectedResourceName string
}

func TestOpentracing(t *testing.T) {
	tracerTests := []opentracingTest{
		// simple path
		{
			routes: func(r *httpx.Router) {
				r.Handle("/path", httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					return nil
				})).Methods("GET")
			},
			req:                  newRequest("GET", "/path"),
			expectedResourceName: "GET /path",
		},
		// path with variables
		{
			routes: func(r *httpx.Router) {
				r.Handle("/users/{user_id}", httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					return nil
				})).Methods("DELETE")
			},
			req:                  newRequest("DELETE", "/users/23"),
			expectedResourceName: "DELETE /users/{user_id}",
		},
		// path with regexp variables
		{
			routes: func(r *httpx.Router) {
				r.Handle("/articles/{category}/{id:[0-9]+}", httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					return nil
				})).Methods("PUT")
			},
			req:                  newRequest("PUT", "/articles/tech/123"),
			expectedResourceName: "PUT /articles/{category}/{id:[0-9]+}",
		},
		// using Path().Handler() style
		{
			routes: func(r *httpx.Router) {
				r.Path("/articles/{category}/{id}").HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					return nil
				}).Methods("GET")
			},
			req:                  newRequest("GET", "/articles/tech/456"),
			expectedResourceName: "GET /articles/{category}/{id}",
		},
		// no route
		{
			routes: func(r *httpx.Router) {
			},
			req:                  newRequest("GET", "/non_existent"),
			expectedResourceName: "GET unknown",
		},
	}

	for _, tt := range tracerTests {
		runOpentracingTest(t, &tt)
	}
}

func runOpentracingTest(t *testing.T, tt *opentracingTest) {
	var m httpx.Handler
	r := httpx.NewRouter()

	if tt.routes != nil {
		tt.routes(r)
	}

	tracer := mocktracer.New()

	opentracing.SetGlobalTracer(tracer)

	m = &OpentracingTracer{
		handler: r,
		router:  r,
	}

	ctx := context.Background()
	resp := httptest.NewRecorder()

	if err := m.ServeHTTPContext(ctx, resp, tt.req); err != nil {
		t.Fatal(err)
	}

	resourceName := tracer.FinishedSpans()[0].Tags()["resource.name"]
	if resourceName != tt.expectedResourceName {
		t.Errorf("Expected %s as resource name, got %s", tt.expectedResourceName, resourceName)
	}
}

func newRequest(method, path string) *http.Request {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}

	return req
}
