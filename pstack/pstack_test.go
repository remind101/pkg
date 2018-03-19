package pstack_test

import (
	"fmt"
	"testing"

	"github.com/remind101/pkg/pstack"
)

func TestNew(t *testing.T) {
	var err error
	defer func() {
		if v := recover(); v != nil {
			var ok bool
			if err, ok = v.(error); ok {
				err = pstack.New(err)
			} else {
				err = pstack.New(fmt.Errorf("%v", v))
			}
		}

		if err == nil {
			t.Fatal("expected err to not be nil")
		}

		if _, ok := err.(error); !ok {
			t.Fatal("expected err to be an error")
		}

		if e, ok := err.(pstack.StackTracer); ok {
			st := e.StackTrace()
			s := fmt.Sprintf("%s:%d", st[0], st[0])
			if got, want := s, "pstack_test.go:42"; got != want {
				t.Errorf("got %s; expected %s", got, want)
			}
		} else {
			t.Fatal("expected err to be a StackTracer")
		}
	}()

	var m []string
	fmt.Println(m[1]) // Will cause a panic
}
