package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/remind101/pkg/client"
	"github.com/remind101/pkg/client/metadata"
)

type mathClient struct {
	c *client.Client
}

type multiplyInput struct {
	A int `json:"a"`
	B int `json:"b"`
}

type multiplyOutput struct {
	Result int `json:"result"`
}

func (mc *mathClient) Multiply(a, b int) (int, error) {
	params := multiplyInput{A: a, B: b}
	var data multiplyOutput

	req := mc.c.NewRequest(context.Background(), "POST", "/multiply", params, &data)
	err := req.Send()

	return data.Result, err
}

func TestClient(t *testing.T) {
	r := mux.NewRouter()
	r.Handle("/multiply", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var params multiplyInput
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			t.Error(err)
		}
		response := multiplyOutput{Result: params.A * params.B}
		json.NewEncoder(rw).Encode(response)
	})).Methods("POST")
	s := httptest.NewServer(r)
	defer s.Close()

	mc := mathClient{
		c: client.New(metadata.ClientInfo{ServiceName: "Math", Endpoint: s.URL}),
	}
	res, err := mc.Multiply(5, 2)
	if err != nil {
		t.Error(err)
	}

	if got, want := res, 10; got != want {
		t.Errorf("got %d; expected %d", got, want)
	}
}
