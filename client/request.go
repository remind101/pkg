package client

import (
	"net/http"
	"time"
)

// Request manages the lifecycle of a client request.
type Request struct {
	Time         time.Time
	HTTPClient   *http.Client
	Handlers     Handlers
	HTTPRequest  *http.Request
	HTTPResponse *http.Response
	Params       interface{} // The input value to encode into the request.
	Data         interface{} // The output value to decode the response into.
	Error        error

	built bool
}

func NewRequest(httpReq *http.Request, handlers Handlers, params interface{}, data interface{}) *Request {
	r := &Request{
		HTTPClient:  http.DefaultClient,
		Handlers:    handlers.Copy(),
		Time:        time.Now(),
		HTTPRequest: httpReq,
		Params:      params,
		Data:        data,
	}

	return r
}

func (r *Request) Send() error {
	defer func() {
		r.Handlers.Complete.Run(r)
	}()

	r.Build()
	r.Sign()
	r.Do()
	r.Decode()

	return r.Error
}

func (r *Request) Build() {
	if !r.built {
		r.Handlers.Build.Run(r)
		r.built = true
	}
}

func (r *Request) Sign() {
	r.Build()
	if r.Error != nil {
		return
	}
	r.Handlers.Sign.Run(r)
}

// Do will perform the request unless an error has already occurred
func (r *Request) Do() {
	if r.Error != nil {
		return
	}
	r.Handlers.Send.Run(r)
}

func (r *Request) Decode() {
	r.Handlers.Decode.Run(r)
}
