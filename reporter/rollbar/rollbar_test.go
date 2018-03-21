package rollbar

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"context"

	"github.com/remind101/pkg/errctx"
	"github.com/remind101/pkg/reporter"
	"github.com/stvp/rollbar"
)

func TestIsAReporter(t *testing.T) {
	var _ reporter.Reporter = &rollbarReporter{}
}

func TestReportsThingsToRollbar(t *testing.T) {
	body := map[string]interface{}{}
	reached := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		body = decodeBody(r.Body)
	}))
	defer ts.Close()

	form := url.Values{}
	form.Add("param1", "param1value")
	req, _ := http.NewRequest("GET", "/api/foo", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	req.Form = form

	boom := fmt.Errorf("boom")
	ctx := context.Background()
	ctx = errctx.WithInfo(ctx, "request_id", "1234")
	ctx = errctx.WithRequest(ctx, req)

	err := errctx.New(ctx, boom, 0)

	ConfigureReporter("token", "test")
	rollbar.Endpoint = ts.URL + "/"
	fmt.Println(ts.URL)
	ctx = reporter.WithReporter(ctx, Reporter)
	reporter.Report(ctx, err)
	rollbar.Wait()

	paramValue := body["data"].(map[string]interface{})["request"].(map[string]interface{})["POST"].(map[string]interface{})["param1"]
	if paramValue != "param1value" {
		t.Fatalf("paramater value didn't make it through to rollbar server")
	}
}

func decodeBody(body io.ReadCloser) map[string]interface{} {
	decoder := json.NewDecoder(body)
	v := map[string]interface{}{}
	err := decoder.Decode(&v)
	if err != nil {
		panic(err)
	}
	return v
}
