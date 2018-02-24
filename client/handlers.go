package client

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	dd_opentracing "github.com/DataDog/dd-trace-go/opentracing"
	opentracing "github.com/opentracing/opentracing-go"
)

type Handlers struct {
	Build    HandlerList
	Sign     HandlerList
	Send     HandlerList
	Decode   HandlerList
	Complete HandlerList
}

func DefaultHandlers() Handlers {
	return Handlers{
		Build:    NewHandlerList(JSONBuilder),
		Sign:     NewHandlerList(),
		Send:     NewHandlerList(BaseSender),
		Decode:   NewHandlerList(JSONDecoder),
		Complete: NewHandlerList(),
	}
}

func (h Handlers) Copy() Handlers {
	return Handlers{
		Build:    h.Build.copy(),
		Sign:     h.Sign.copy(),
		Send:     h.Send.copy(),
		Decode:   h.Decode.copy(),
		Complete: h.Complete.copy(),
	}
}

type HandlerList struct {
	list []Handler
}

func NewHandlerList(hh ...Handler) HandlerList {
	return HandlerList{
		list: append([]Handler{}, hh...),
	}
}

func (hl *HandlerList) Run(r *Request) {
	for _, h := range hl.list {
		h.Fn(r)
	}
}

func (hl *HandlerList) Append(h Handler) {
	hl.list = append(hl.list, h)
}

func (hl *HandlerList) copy() HandlerList {
	n := HandlerList{}
	if len(hl.list) == 0 {
		return n
	}

	n.list = append(make([]Handler, 0, len(hl.list)), hl.list...)
	return n
}

type Handler struct {
	Name string
	Fn   func(*Request)
}

// BaseSender sends a request using the http.Client.
var BaseSender = Handler{
	Name: "BaseSender",
	Fn: func(r *Request) {
		r.HTTPResponse, r.Error = r.HTTPClient.Do(r.HTTPRequest)
	},
}

var JSONBuilder = Handler{
	Name: "JSONBuilder",
	Fn: func(r *Request) {
		r.HTTPRequest.Header.Set("Content-Type", "application/json")
		r.HTTPRequest.Header.Set("Accept", "application/json")

		if r.HTTPRequest.Method != "GET" && r.Params != nil {
			raw, err := json.Marshal(r.Params)
			if err != nil {
				r.Error = err
				return
			}
			r.HTTPRequest.Body = ioutil.NopCloser(bytes.NewReader(raw))
		}
	},
}

// JSONDecoder decodes a response as JSON.
var JSONDecoder = Handler{
	Name: "JSONDecoder",
	Fn: func(r *Request) {
		if r.HTTPResponse == nil {
			return
		}
		if r.HTTPResponse.Body != nil {
			defer r.HTTPResponse.Body.Close()
		}
		if r.Data == nil {
			_, r.Error = io.Copy(ioutil.Discard, r.HTTPResponse.Body)
			return
		}
		r.Error = json.NewDecoder(r.HTTPResponse.Body).Decode(r.Data)
	},
}

// BasicAuther sets basic auth on a request.
func BasicAuther(username, password string) Handler {
	return Handler{
		Name: "BasicAuther",
		Fn: func(r *Request) {
			r.HTTPRequest.SetBasicAuth(username, password)
		},
	}
}

// WithTracing returns a Send Handler that wraps another Send Handler in a trace
// span.
func WithTracing(h Handler, r *Request) Handler {
	return Handler{
		Name: "TracedSender",
		Fn: func(r *Request) {
			span, ctx := opentracing.StartSpanFromContext(r.HTTPRequest.Context(), "client.request")
			defer span.Finish()
			r.HTTPRequest = r.HTTPRequest.WithContext(ctx)

			h.Fn(r)

			if r.Error != nil {
				span.SetTag(dd_opentracing.Error, r.Error)
			}
		},
	}
}