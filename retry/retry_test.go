package retry

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

type MyError struct{}

func (e *MyError) Error() string {
	return "Error"
}

type Counter struct {
	sync.Mutex
	count int
}

func (c *Counter) Incr() {
	c.Lock()
	defer c.Unlock()
	c.count++
}

func (c *Counter) Count() int {
	return c.count
}

var alwaysRetry = func(error) bool { return true }
var neverRetry = func(error) bool { return false }

func createRetrier(shouldRetry func(error) bool) (*Retrier, *Counter) {
	backOffOpts := &BackOffOpts{
		InitialInterval: 1 * time.Nanosecond,
		MaxInterval:     5 * time.Nanosecond,
		MaxElapsedTime:  250 * time.Microsecond}
	counter := &Counter{}

	return NewRetrier("Retrier", backOffOpts, shouldRetry), counter
}

func TestRetriesUntilMaxElapsedTimeAndCallsNotifyGaveUp(t *testing.T) {
	retrier, counter := createRetrier(alwaysRetry)

	notifyGaveUpCalled := 0
	retrier.AddNotifyGaveUp(func(*RetryEvent) { notifyGaveUpCalled++ })

	_, err := retrier.Retry(func() (interface{}, error) {
		counter.Incr()
		return 0, &MyError{}
	})

	want := &MyError{}
	if err != want {
		t.Fatalf("Expected %v, got %v", err, want)
	}
	if counter.Count() < 2 {
		t.Fatalf("Expected %v >= 2", counter.Count())
	}
	if notifyGaveUpCalled != 1 {
		t.Fatalf("Expected notifygaveup to have been called")
	}
}

func TestRetriesUntilSuccessfulCallingNotifyRetryOnEachRetry(t *testing.T) {
	retrier, counter := createRetrier(alwaysRetry)

	notifyRetryCalled := 0
	retrier.AddNotifyRetry(func(*RetryEvent) { notifyRetryCalled++ })

	iVal, err := retrier.Retry(func() (interface{}, error) {
		counter.Incr()
		if counter.Count() < 5 {
			return 0, &MyError{}
		} else {
			return 123, nil
		}
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	val := iVal.(int)

	// We retried 4 times, for a total of 5 tries.
	if counter.Count() != 5 && val != 123 && notifyRetryCalled != 4 {
		t.Fatalf("Unexpected Counter: %v, val %v, notifyRetryCalled: %v", counter.Count(), val, notifyRetryCalled)
	}
}

func TestReturnsTheErrorForNonRetryableErrors(t *testing.T) {
	retrier, _ := createRetrier(neverRetry)
	myError := errors.New("myError")
	_, err := retrier.Retry(func() (interface{}, error) {
		return nil, myError
	})
	if err != myError {
		t.Fatalf("Expected to receive: %v, got %v", myError, err)
	}
}

func TestRetryWhenErrorTypeMatches(t *testing.T) {
	var MyErrorType = reflect.TypeOf(&MyError{})

	shouldRetryFunc := RetryWhenErrorTypeMatches([]reflect.Type{MyErrorType})

	if got := shouldRetryFunc(&MyError{}); got != true {
		t.Fatalf("Expected true")
	}
	if got := shouldRetryFunc(errors.New("hi")); got != false {
		t.Fatalf("Expected false")
	}
}
