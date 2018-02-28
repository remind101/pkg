package request

import (
	"net/http"
	"time"
)

// Request manages the lifecycle of a client request.
type Request struct {
	Time         time.Time      // The time the request was created.
	HTTPClient   *http.Client   // The underlying http.Client that will make the request.
	Handlers     Handlers       // Handlers contains the logic that manages the lifecycle of the request.
	HTTPRequest  *http.Request  // The http.Request object
	HTTPResponse *http.Response // The http.Response object, should be populated after Handlers.Send has run.
	Params       interface{}    // The input value to encode into the request.
	Data         interface{}    // The output value to decode the response into.
	Error        error          // Holds any error that occurs during request sending.

	built bool // True if request has been built already.
}

// New creates a new Request.
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

// Send sends a request. Send manages the execution of the Handlers.
// If an error occurs during any phase, request processing stops.
func (r *Request) Send() error {
	// Always run Complete handlers
	defer func() {
		r.Handlers.Complete.Run(r)
	}()

	// Build and Sign the request.
	r.Build()
	if r.Error != nil {
		return r.Error
	}

	// Send the request. r.HTTPResponse should be populated after these handlers run.
	r.Handlers.Send.Run(r)
	if r.Error != nil {
		return r.Error
	}

	// Validate the response. This is a good place to create a r.Error if
	// the response is not a 2xx status code.
	r.Handlers.ValidateResponse.Run(r)
	if r.Error != nil {
		// Run DecodeError handlers. This is a good place to add custom error response
		// parsing.
		r.Handlers.DecodeError.Run(r)
		return r.Error
	}

	// Decode the response. Commonly will populate r.Data variable with a parsed response body.
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
