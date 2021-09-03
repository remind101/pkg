package rollbar

import (
	"net/http"
	"os"
	"strconv"

	"context"

	"github.com/remind101/pkg/reporter/util"
	"github.com/rollbar/rollbar-go"
)

const ErrorLevel = "error"

const (
	EnvAccessToken = "ROLLBAR_ACCESS_TOKEN"
	EnvEnvironment = "ROLLBAR_ENVIRONMENT"
	EnvEndpoint    = "ROLLBAR_ENDPOINT"
)

// TODO (sophied): A type temporarily stolen from github.com/stvp/rollbar
// to allow us to migrate all the places that use this reporting package
// away from a custom stack trace type before we switch over to go's default
// error type across the board.
type Frame struct {
	Filename string `json:"filename"`
	Method   string `json:"method"`
	Line     int    `json:"lineno"`
}

type rollbarReporter struct{}

var Reporter = &rollbarReporter{}

func ConfigureReporter(token, environment string) {
	rollbar.SetToken(token)
	rollbar.SetEnvironment(environment)
}

func ConfigureFromEnvironment() {
	if token := os.Getenv(EnvAccessToken); token != "" {
		rollbar.SetToken(token)
	}
	if env := os.Getenv(EnvEnvironment); env != "" {
		rollbar.SetEnvironment(env)
	}
	if endpoint := os.Getenv(EnvEndpoint); endpoint != "" {
		rollbar.SetEndpoint(endpoint)
	}
}

func (r *rollbarReporter) ReportWithLevel(ctx context.Context, level string, err error) error {
	var request *http.Request
	var extraFields map[string]interface{}

	if e, ok := err.(util.Contexter); ok {
		extraFields = e.ContextData()
	}

	if e, ok := err.(util.Requester); ok {
		request = e.Request()
	}

	if e, ok := err.(util.Causer); ok {
		err = e.Cause() // Report the actual cause of the error.
	}

	reportToRollbar(level, request, err, extraFields)
	return nil
}

func (r *rollbarReporter) Flush() {
	rollbar.Wait()
}

func reportToRollbar(level string, request *http.Request, err error, extraFields map[string]interface{}) {
	if request != nil {
		rollbar.Error(level, request, err, extraFields)
	} else {
		rollbar.Error(level, err, extraFields)
	}
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
