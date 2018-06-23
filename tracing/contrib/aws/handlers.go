package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	opentracing "github.com/opentracing/opentracing-go"
)

var (
	// Starts a span and adds it to the request context.
	StartHandler = request.NamedHandler{
		Name: "opentracing.Start",
		Fn: func(r *request.Request) {
			_, ctx := opentracing.StartSpanFromContext(r.Context(), "client.request")
			r.SetContext(ctx)
		},
	}

	// Adds information about the request to the span.
	RequestInfoHandler = request.NamedHandler{
		Name: "opentracing.RequestInfo",
		Fn: func(r *request.Request) {
			span := opentracing.SpanFromContext(r.Context())
			span.SetTag("service.name", fmt.Sprintf("aws.%s", r.ClientInfo.ServiceName))
			span.SetTag("resource.name", r.Operation.Name)
			span.SetTag("http.method", r.Operation.HTTPMethod)
			span.SetTag("http.url", r.ClientInfo.Endpoint+r.Operation.HTTPPath)
			span.SetTag("out.host", r.ClientInfo.Endpoint)
			span.SetTag("aws.operation", r.Operation.Name)
		},
	}

	// Finishes the span.
	FinishHandler = request.NamedHandler{
		Name: "opentracing.Finish",
		Fn: func(r *request.Request) {
			span := opentracing.SpanFromContext(r.Context())
			span.SetTag("aws.retry_count", fmt.Sprintf("%d", r.RetryCount))

			if r.HTTPResponse != nil {
				span.SetTag("http.status_code", fmt.Sprintf("%d", r.HTTPResponse.StatusCode))
			}

			if r.Error != nil {
				span.SetTag("error.error", r.Error)
				if err, ok := r.Error.(awserr.Error); ok {
					span.SetTag("aws.err.code", fmt.Sprintf("%s", err.Code()))
				}
			}

			span.Finish()
		},
	}
)

// WithTracing adds the necessary request handlers to an AWS session.Session
// object to enable tracing with opentracing.
func WithTracing(s *session.Session) {
	// After adding these handlers, the "Send" handler list will look
	// something like:
	//
	//	opentracing.Start -> opentracing.RequestInfo -> core.ValidateReqSigHandler -> core.SendHandler
	s.Handlers.Send.PushFrontNamed(RequestInfoHandler)
	s.Handlers.Send.PushFrontNamed(StartHandler)

	s.Handlers.Complete.PushBackNamed(FinishHandler)
}
