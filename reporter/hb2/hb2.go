// package hb2 is a Go package for sending errors to Honeybadger
// using the official client library
package hb2

import (
	"fmt"
	"net/http"
	"strings"

	"context"

	"github.com/pkg/errors"
	"github.com/remind101/pkg/reporter/hb2/internal/honeybadger-go"
	"github.com/remind101/pkg/reporter/util"
)

// Headers that won't be sent to honeybadger.
var IgnoredHeaders = map[string]struct{}{
	"Authorization": struct{}{},
}

type Config struct {
	ApiKey      string
	Environment string
	Endpoint    string
}

type HbReporter struct {
	client *honeybadger.Client
}

// NewReporter returns a new Reporter instance.
func NewReporter(cfg Config) *HbReporter {
	hbCfg := honeybadger.Configuration{}
	hbCfg.APIKey = cfg.ApiKey
	hbCfg.Env = cfg.Environment
	hbCfg.Endpoint = cfg.Endpoint

	return &HbReporter{honeybadger.New(hbCfg)}
}

// exposes honeybadger config for unit tests
func (r *HbReporter) GetConfig() *honeybadger.Configuration {
	return r.client.Config
}

// ReportWithLevel reports reports the error to honeybadger.
func (r *HbReporter) ReportWithLevel(ctx context.Context, level string, err error) error {
	extras := []interface{}{}

	if e, ok := err.(util.Contexter); ok {
		extras = append(extras, getContextData(e))
	}

	if e, ok := err.(util.Requester); ok {
		if r := e.Request(); r != nil {
			extras = append(extras, honeybadger.Params(r.Form), getRequestData(r), *r.URL)
		}
	}

	err = makeHoneybadgerError(err)

	_, clientErr := r.client.Notify(err, extras...)
	return clientErr
}

func (r *HbReporter) Flush() {
	r.client.Flush()
}

func getRequestData(r *http.Request) honeybadger.CGIData {
	cgiData := honeybadger.CGIData{}
	replacer := strings.NewReplacer("-", "_")

	for header, values := range r.Header {
		if _, ok := IgnoredHeaders[header]; ok {
			continue
		}
		key := "HTTP_" + replacer.Replace(strings.ToUpper(header))
		cgiData[key] = strings.Join(values, ",")
	}

	cgiData["REQUEST_METHOD"] = r.Method
	return cgiData
}

func getContextData(err util.Contexter) honeybadger.Context {
	ctx := honeybadger.Context{}
	for key, value := range err.ContextData() {
		ctx[key] = value
	}
	return ctx
}

func makeHoneybadgerError(err error) honeybadger.Error {
	className := util.ClassName(err)
	if e, ok := err.(util.Causer); ok {
		className = util.ClassName(e.Cause())
	}

	frames := make([]*honeybadger.Frame, 0)
	if e, ok := err.(util.StackTracer); ok {
		frames = makeHoneybadgerFrames(e.StackTrace())
	}

	return honeybadger.Error{
		Message: err.Error(),
		Class:   className,
		Stack:   frames,
	}
}

func makeHoneybadgerFrames(stack errors.StackTrace) []*honeybadger.Frame {
	length := len(stack)
	frames := make([]*honeybadger.Frame, length)
	for index, frame := range stack[:length] {
		frames[index] = &honeybadger.Frame{
			Number: fmt.Sprintf("%d", frame),
			File:   fmt.Sprintf("%s", frame),
			Method: fmt.Sprintf("%n", frame),
		}
	}
	return frames
}
