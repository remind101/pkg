package aws

import (
	"fmt"

	dd_ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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
				span.SetTag(dd_ext.Error, r.Error)
				if _, ok := r.Error.(fmt.Formatter); ok {
					span.SetTag(dd_ext.ErrorStack, fmt.Sprintf("%+v", r.Error))
				}
				if err, ok := r.Error.(awserr.Error); ok {
					span.SetTag("aws.err.code", fmt.Sprintf("%s", err.Code()))
				}
			}

			span.Finish()
		},
	}

	// DynamoDBInfoHandler adds extra information to the span for requests
	// to DynamoDB.
	DynamoDBInfoHandler = request.NamedHandler{
		Name: "opentracing.DynamoDB",
		Fn: func(r *request.Request) {
			// Check if this request is for the DynamoDB API.
			if r.ClientInfo.ServiceName != dynamodb.ServiceName {
				return
			}

			tableName := dynamoDBTableName(r)
			if tableName == nil {
				return
			}

			span := opentracing.SpanFromContext(r.Context())
			span.SetOperationName(r.Operation.Name)
			span.SetTag("span.type", "db")
			span.SetTag("resource.name", *tableName)
			span.SetTag("aws.dynamodb.table", *tableName)
		},
	}
)

// WithTracing adds the necessary request handlers to an AWS session.Session
// object to enable tracing with opentracing.
func WithTracing(s *session.Session) {
	// After adding these handlers, the "Send" handler list will look
	// something like:
	//
	//	opentracing.Start -> opentracing.RequestInfo -> opentracing.DynamoDBInfo -> core.ValidateReqSigHandler -> core.SendHandler
	s.Handlers.Send.PushFrontNamed(DynamoDBInfoHandler)
	s.Handlers.Send.PushFrontNamed(RequestInfoHandler)
	s.Handlers.Send.PushFrontNamed(StartHandler)

	s.Handlers.Complete.PushBackNamed(FinishHandler)
}

// dynamoDBTableName attempts to return the name of the DynamoDB table that this
// request is operating on.
func dynamoDBTableName(r *request.Request) *string {
	// All DynamoDB requests that operate on a table.
	switch v := r.Params.(type) {
	case *dynamodb.QueryInput:
		return v.TableName
	case *dynamodb.ScanInput:
		return v.TableName

	case *dynamodb.PutItemInput:
		return v.TableName
	case *dynamodb.GetItemInput:
		return v.TableName
	case *dynamodb.UpdateItemInput:
		return v.TableName
	case *dynamodb.DeleteItemInput:
		return v.TableName

	case *dynamodb.CreateTableInput:
		return v.TableName
	case *dynamodb.UpdateTableInput:
		return v.TableName
	case *dynamodb.DeleteTableInput:
		return v.TableName

	case *dynamodb.CreateBackupInput:
		return v.TableName
	case *dynamodb.ListBackupsInput:
		return v.TableName
	case *dynamodb.DescribeContinuousBackupsInput:
		return v.TableName
	case *dynamodb.UpdateContinuousBackupsInput:
		return v.TableName

	case *dynamodb.DescribeTableInput:
		return v.TableName
	case *dynamodb.DescribeTimeToLiveInput:
		return v.TableName
	default:
		return nil
	}
}
