package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/remind101/pkg/client"
)

type mathClient struct {
	c *client.Client
}

type muliplyInput struct {
	A int `json:"a"`
	B int `json:"b"`
}

type muliplyOutput struct {
	Result int `json:"result"`
}

func (mc *mathClient) Multiply(a, b int) (int, error) {
	params := muliplyInput{A: a, B: b}
	var data muliplyOutput

	req := mc.c.NewRequest(context.Background(), "POST", "/multiply", params, &data)
	err := req.Send()

	return data.Result, err
}

func TestClient(t *testing.T) {
	r := mux.NewRouter()
	r.Handle("/multiply", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var params muliplyInput
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			t.Error(err)
		}
		response := muliplyOutput{Result: params.A * params.B}
		json.NewEncoder(rw).Encode(response)
	})).Methods("POST")
	s := httptest.NewServer(r)
	defer s.Close()

	mc := mathClient{client.New(s.URL)}
	res, err := mc.Multiply(5, 2)
	if err != nil {
		t.Error(err)
	}

	if got, want := res, 10; got != want {
		t.Errorf("got %d; expected %d", got, want)
	}
}
