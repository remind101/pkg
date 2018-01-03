package httpx

import "golang.org/x/net/context"

func WithHeader(ctx context.Context, headerKey, headerValue string) context.Context {
	return context.WithValue(ctx, headerKey, headerValue)
}

func Header(ctx context.Context, headerKey string) string {
	headerValue, _ := ctx.Value(headerKey).(string)
	return headerValue
}
