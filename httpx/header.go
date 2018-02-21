package httpx

import "context"

type headerContextKey string

func (c headerContextKey) String() string {
	return "pkg header " + string(c)
}

func WithHeader(ctx context.Context, headerKey string, headerValue string) context.Context {
	return context.WithValue(ctx, headerContextKey(headerKey).String(), headerValue)
}

func Header(ctx context.Context, headerKey string) string {
	headerValue, _ := ctx.Value(headerContextKey(headerKey).String()).(string)
	return headerValue
}
