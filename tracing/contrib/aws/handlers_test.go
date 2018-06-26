package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/mock"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	awsot "github.com/remind101/pkg/tracing/contrib/aws"
)

func TestWithTracing(t *testing.T) {
	tracer := mocktracer.New()
	opentracing.SetGlobalTracer(tracer)

	awsot.WithTracing(mock.Session)

	_, ctx := opentracing.StartSpanFromContext(context.Background(), "root")

	client := mock.NewMockClient(&aws.Config{Region: aws.String("us-west-2")})

	r := client.NewRequest(&request.Operation{
		Name:       "object.get",
		HTTPMethod: "GET",
		HTTPPath:   "/foobar",
	}, nil, nil)

	r.SetContext(ctx)
	err := r.Send()
	if err != nil {
		t.Fatal(err)
	}

	spans := tracer.FinishedSpans()
	if len(spans) != 1 {
		t.Fatal("expected 1 finished span")
	}
	span := spans[0]

	if got, want := span.OperationName, "client.request"; got != want {
		t.Errorf("got: %+v; expected %+v", got, want)
	}

	assertTags(t, span, map[string]string{
		"service.name":     "aws.Mock",
		"resource.name":    "object.get",
		"http.method":      "GET",
		"http.url":         client.Endpoint + "/foobar",
		"http.status_code": "200",
		"out.host":         client.Endpoint,
		"aws.operation":    "object.get",
		"aws.retry_count":  "0",
	})
}

func assertTags(t testing.TB, span *mocktracer.MockSpan, tags map[string]string) {
	for k, v := range tags {
		tagName, ok := span.Tag(k).(string)
		if !ok {
			t.Errorf("no %s tag", k)
		}

		if got, want := tagName, v; got != want {
			t.Errorf("span.Tag('%s'): %+v; expected %+v", k, got, want)
		}
	}
}
