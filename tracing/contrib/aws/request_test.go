package aws_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/mock"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	awsot "github.com/remind101/pkg/tracing/contrib/aws"
)

func TestSend(t *testing.T) {
	tracer := mocktracer.New()
	opentracing.SetGlobalTracer(tracer)

	_, ctx := opentracing.StartSpanFromContext(context.Background(), "root")

	client := mock.NewMockClient(&aws.Config{Region: aws.String("us-west-2")})
	r := client.NewRequest(&request.Operation{
		Name:       "object.get",
		HTTPMethod: "GET",
		HTTPPath:   "/foobar",
	}, nil, nil)

	awsTracer := awsot.New()

	err := awsTracer.Send(ctx, r, func(s opentracing.Span, r *request.Request, err error) {
		s.SetTag("span.type", "external")
		s.SetTag("service.name", "aws.s3")
	})

	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}

	spans := tracer.FinishedSpans()
	if len(spans) != 1 {
		t.Fatal("expected 1 finished span")
	}
	span := spans[0]
	if got, want := span.OperationName, "client.request"; got != want {
		t.Errorf("got: %+v; expected %+v", got, want)
	}

	tags := map[string]string{
		"span.type":     "external",
		"service.name":  "aws.s3",
		"resource.name": "object.get",
		"http.method":   "GET",
		"http.url":      client.Endpoint + "/foobar",
		"out.host":      client.Endpoint,
		"aws.operation": "object.get",
	}

	for k, v := range tags {
		if got, want := span.Tag(k).(string), v; got != want {
			t.Errorf("span.Tag('%s'): %+v; expected %+v", k, got, want)
		}
	}

	if got, want := span.Tag("http.status_code").(int), http.StatusOK; got != want {
		t.Errorf("got: %+v; expected %+v", got, want)
	}

	if got, want := span.Tag("aws.retry_count").(int), 0; got != want {
		t.Errorf("got: %+v; expected %+v", got, want)
	}
}
