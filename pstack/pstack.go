package pstack

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

type StackTracer interface {
	error
	StackTrace() errors.StackTrace
}

type errorWithStack struct {
	Err   error
	Stack errors.StackTrace
}

func (e *errorWithStack) Error() string {
	return e.Err.Error()
}

func (e *errorWithStack) StackTrace() errors.StackTrace {
	return e.Stack
}

// New will convert a stack trace gathered from a panic to an errorWithStack
func New(v interface{}) (e error) {
	switch err := v.(type) {
	case StackTracer:
		e = err.(error)
	case error:
		s := callers()
		e = &errorWithStack{
			Err:   err,
			Stack: s.StackTrace(),
		}
	default:
		s := callers()
		e = &errorWithStack{
			Err:   fmt.Errorf("%v", err),
			Stack: s.StackTrace(),
		}
	}

	return e
}

// stack represents a stack of program counters.
type stack []uintptr

func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := errors.Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

func (s *stack) StackTrace() errors.StackTrace {
	f := make([]errors.Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = errors.Frame((*s)[i])
	}
	return f
}

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(4, pcs[:])

	var si int
	for i, pc := range pcs {
		si = i
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		if !strings.HasPrefix(fn.Name(), "runtime.") {
			break
		}
	}

	var st stack = pcs[si:n]
	return &st
}
