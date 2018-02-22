package service_client_test

import (
	"context"
	"fmt"

	"github.com/remind101/pkg/service_client"
)

func Example() {
	client := service_client.NewServiceClient("http://example.org")

	response := struct {
		Error string `json:"error"`
	}{}

	// If this context contains a request trace, the client will add a client.request
	// span to it.
	ctx := context.Background()

	// Simple GET request
	err := client.Do(ctx, "GET", "/path", nil, &response)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Simple POST request
	jsonData := &struct {
		Message string `json:"message"`
	}{
		Message: "Hello",
	}
	err = client.Do(ctx, "POST", "/path", jsonData, &response)
	if err != nil {
		fmt.Println(err.Error())
	}
}
