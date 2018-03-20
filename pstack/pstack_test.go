package pstack_test

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/remind101/pkg/pstack"
)

type panicTest struct {
	Fn     func()
	TestFn func(error)
}

func TestPanics(t *testing.T) {
	tests := []panicTest{
		{
			Fn: func() {},
			TestFn: func(err error) {
				if err != nil {
					t.Error("expected err to be nil")
				}
			},
		},
		{
			Fn: func() {
				panic("boom!")
			},
			TestFn: func(err error) {
				if err == nil {
					t.Error("expected err to not be nil")
				}
				if e, ok := err.(pstack.StackTracer); ok {
					if got, want := fmt.Sprintf("%v", e.StackTrace()[0]), "pstack_test.go:29"; got != want {
						t.Errorf("got: %v; expected: %v", got, want)
					}
				}
			},
		},
		{
			Fn: func() {
				panic(fmt.Errorf("boom!"))
			},
			TestFn: func(err error) {
				if err == nil {
					t.Error("expected err to not be nil")
				}
				if e, ok := err.(pstack.StackTracer); ok {
					if got, want := fmt.Sprintf("%v", e.StackTrace()[0]), "pstack_test.go:44"; got != want {
						t.Errorf("got: %v; expected: %v", got, want)
					}
				}
			},
		},
		{
			Fn: func() {
				panic(errors.New("boom"))
			},
			TestFn: func(err error) {
				if err == nil {
					t.Error("expected err to not be nil")
				}
				if e, ok := err.(pstack.StackTracer); ok {
					if got, want := fmt.Sprintf("%v", e.StackTrace()[0]), "pstack_test.go:59"; got != want {
						t.Errorf("got: %v; expected: %v", got, want)
					}
				}
			},
		},
	}
	for _, tt := range tests {
		runPanicTest(tt)
	}
}

func runPanicTest(pt panicTest) {
	defer func() {
		err := pstack.New(recover())
		pt.TestFn(err)
	}()

	pt.Fn()
}
