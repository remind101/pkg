package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/remind101/newrelic"
	"github.com/stretchr/testify/mock"
)

func TestNewRelicTracer(t *testing.T) {
	var called bool

	tx := new(mockTx)
	h := &NewRelicTracer{
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			_, ok := newrelic.FromContext(r.Context())
			if !ok {
				t.Fatal("expected transaction to be in context")
			}
		}),
		newTx: func(*http.Request) newrelic.Tx {
			return tx
		},
	}

	tx.On("Start").Return(nil)
	tx.On("End").Return(nil)

	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()

	h.ServeHTTP(resp, req)

	if !called {
		t.Fatal("expected handler to be called")
	}
}

type mockTx struct {
	newrelic.Tx
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
