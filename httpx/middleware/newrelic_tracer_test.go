package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/remind101/nra"
	"github.com/remind101/pkg/httpx"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

type tracerTest struct {
	// A function that adds Handlers to the router.
	routes func(*httpx.Router)

	// An http.Request to test.
	req *http.Request

	expectedTransactionName string
	expectedUrl             string
}

func TestTracing(t *testing.T) {
	tracerTests := []tracerTest{
		// simple path
		{
			routes: func(r *httpx.Router) {
				r.Handle("/path", httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					return nil
				})).Methods("GET")
			},
			req: newRequest("GET", "/path"),
			expectedTransactionName: "/path (GET)",
			expectedUrl:             "/path",
		},
		// path with variables
		{
			routes: func(r *httpx.Router) {
				r.Handle("/users/{user_id}", httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					return nil
				})).Methods("DELETE")
			},
			req: newRequest("DELETE", "/users/23"),
			expectedTransactionName: "/users/:user_id (DELETE)",
			expectedUrl:             "/users/23",
		},
		// no route
		{
			routes: func(r *httpx.Router) {
			},
			req: newRequest("GET", "/non_existent"),
			expectedTransactionName: "/non_existent (GET)",
			expectedUrl:             "/non_existent",
		},
	}

	for _, tt := range tracerTests {
		traceTest(t, &tt)
	}
}

func traceTest(t *testing.T, tt *tracerTest) {
	var m httpx.Handler
	r := httpx.NewRouter()

	if tt.routes != nil {
		tt.routes(r)
	}

	tx := new(mockTx)
	m = &NewRelicTracer{
		handler: r,
		router:  r,
		tracer:  nil,
		createTx: func(transactionName, url string, tracer nra.TxTracer) nra.Tx {
			if tt.expectedTransactionName != transactionName {
				t.Fatalf("Transaction mismatch expected: %v got: %v", tt.expectedTransactionName, transactionName)
			}
			if tt.expectedUrl != url {
				t.Fatalf("Url mismatch expected: %v got: %v", tt.expectedUrl, url)
			}
			return tx
		},
	}

	ctx := context.Background()
	resp := httptest.NewRecorder()

	tx.On("Start").Return(nil)
	tx.On("End").Return(nil)

	if err := m.ServeHTTPContext(ctx, resp, tt.req); err != nil {
		t.Fatal(err)
	}

	tx.AssertExpectations(t)
}

func newRequest(method, path string) *http.Request {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}

	return req
}

type mockTx struct {
	mock.Mock
}

func (t *mockTx) Start() error {
	args := t.Called()
	return args.Error(0)
}

func (t *mockTx) End() error {
	args := t.Called()
	return args.Error(0)
}

func (t *mockTx) StartGeneric(name string) {
	t.Called(name)
	return
}

func (t *mockTx) StartDatastore(table, operation, sql, rollupName string) {
	t.Called(table, operation, sql, rollupName)
	return
}

func (t *mockTx) StartExternal(host, name string) {
	t.Called(host, name)
	return
}

func (t *mockTx) EndSegment() error {
	args := t.Called()
	return args.Error(0)
}
