package request

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

func New(httpReq *http.Request, handlers Handlers, params interface{}, data interface{}) *Request {
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
	if r.Error != nil {
		return r.Error
	}

	r.Handlers.Send.Run(r)
	if r.Error != nil {
		return r.Error
	}

	r.Handlers.ValidateResponse.Run(r)
	if r.Error != nil {
		r.Handlers.DecodeError.Run(r)
		return r.Error
	}

	r.Handlers.Decode.Run(r)
	return r.Error
}

// Build runs build handlers and then runs sign handlers.
func (r *Request) Build() {
	if !r.built {
		r.Handlers.Build.Run(r)
		r.built = true
		if r.Error != nil {
			return
		}
		r.Handlers.Sign.Run(r)
	}
}
