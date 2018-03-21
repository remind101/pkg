package util

import (
	"net/http"
	"reflect"
	"runtime"

	"github.com/pkg/errors"
)

func ClassName(err error) string {
	return reflect.TypeOf(err).String()
}

func FunctionName(pc uintptr) string {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "???"
	}
	return fn.Name()
}

type StackTracer interface {
	StackTrace() errors.StackTrace
}

type Requester interface {
	Request() *http.Request
}

type Causer interface {
	Cause() error
}

type Contexter interface {
	ContextData() map[string]interface{}
}
