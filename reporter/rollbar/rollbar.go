package rollbar

import (
	"fmt"
	"net/http"
	"strconv"

	"context"

	"github.com/pkg/errors"
	"github.com/remind101/pkg/errctx"
	"github.com/stvp/rollbar"
)

const ErrorLevel = "error"

type rollbarReporter struct{}

// The stvp/rollbar package is implemented as a global, so let's not fool our
// callers by generating an instance of a reporter. Rollbar config is actually
// global, so we'll have the Rollbar reporter be a global too.
var Reporter = &rollbarReporter{}

func ConfigureReporter(token, environment string) {
	rollbar.Token = token
	rollbar.Environment = environment
}

func (r *rollbarReporter) ReportWithLevel(ctx context.Context, level string, err error) error {
	var request *http.Request
	extraFields := []*rollbar.Field{}
	var stackTrace rollbar.Stack = nil

	if e, ok := err.(*errctx.Error); ok {
		extraFields = getContextData(e)

		if r := e.Request; r != nil {
			request = e.Request
		}

		stackTrace = makeRollbarStack(e.StackTrace())
		err = e.Cause() // Report the actual cause of the error.
	}

	reportToRollbar(level, request, err, stackTrace, extraFields)
	return nil
}

func reportToRollbar(level string, request *http.Request, err error, stack rollbar.Stack, extraFields []*rollbar.Field) {
	if request != nil {
		if stack != nil {
			rollbar.RequestErrorWithStack(level, request, err, stack, extraFields...)
		} else {
			rollbar.RequestError(level, request, err, extraFields...)
		}
	} else {
		if stack != nil {
			rollbar.ErrorWithStack(level, err, stack, extraFields...)
		} else {
			rollbar.Error(level, err, extraFields...)
		}
	}
}

func makeRollbarStack(stack errors.StackTrace) rollbar.Stack {
	length := len(stack)
	rollbarStack := make(rollbar.Stack, length)
	for index, frame := range stack[:length] {
		// Rollbar's website has a "most recent call last" header. We need to
		// reverse the order of the stack frames we send it, so our stack traces
		// are shown in that order.
		rollbarStack[length-index-1] = rollbar.Frame{
			Line:     parseInt(fmt.Sprintf("%d", frame)),
			Filename: fmt.Sprintf("%s", frame),
			Method:   fmt.Sprintf("%n", frame)}
	}
	return rollbarStack
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func getContextData(err *errctx.Error) []*rollbar.Field {
	fields := []*rollbar.Field{}
	for key, value := range err.Context {
		fields = append(fields, &rollbar.Field{key, value})
	}
	return fields
}
