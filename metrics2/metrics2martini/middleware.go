package metrics2martini

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/remind101/pkg/metrics2"
)

// ResponseTimeReporter reports timing metrics using metrics2 package
//
// Usage:
//   r := martini.NewRouter()
//   r.Get("/boom/error",
// 	ResponseTimeReporter(),
// 	func(req *http.Request) {
//        ...
// 	})
//
// It is important to insert it after routing, not a a generic martini middleware!
//
func ResponseTimeReporter() martini.Handler {
	return func(res http.ResponseWriter, c martini.Context, r martini.Route) {
		t := metrics2.ResponseTime()
		defer t.Done()

		rw := res.(martini.ResponseWriter)
		c.Next()

		t.SetTags(map[string]string{
			"route":  fmt.Sprintf("%s %s", r.Method(), r.Pattern()),
			"status": strconv.Itoa(rw.Status()),
		})
	}
}
