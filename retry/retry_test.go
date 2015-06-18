package retry

import (
	"errors"
	"reflect"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

var _ = Describe("Retrier", func() {
	var backOffOpts *BackOffOpts
	var counter *Counter

	BeforeEach(func() {
		backOffOpts = &BackOffOpts{
			InitialInterval: 1 * time.Nanosecond,
			MaxInterval:     5 * time.Nanosecond,
			MaxElapsedTime:  250 * time.Microsecond}
		counter = &Counter{}
	})

	It("keeps retrying until MaxElapsedTime and calls NotifyGaveUp", func() {
		retrier := NewRetrier("Retrier", backOffOpts, func(error) bool {
			return true
		})

		notifyGaveUpCalled := 0
		retrier.AddNotifyGaveUp(func(*RetryEvent) { notifyGaveUpCalled++ })

		_, err := retrier.Retry(func() (interface{}, error) {
			counter.Incr()
			return 0, &MyError{}
		})

		Expect(err).To(Equal(&MyError{}))

		Expect(counter.Count()).To(BeNumerically(">", 10))
		Expect(notifyGaveUpCalled).To(Equal(1))
	})

	It("retries until successful, calling NotifyRetry on each retry", func() {
		retrier := NewRetrier("Retrier", backOffOpts, func(error) bool {
			return true
		})

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
		Expect(err).NotTo(HaveOccurred())

		val := iVal.(int)
		Expect(counter.Count()).To(Equal(5))
		Expect(val).To(Equal(123))

		// We retried 4 times, for a total of 5 tries.
		Expect(notifyRetryCalled).To(Equal(4))
	})

	It("returns the error when there's a non-retryable error", func() {
		retrier := NewRetrier("Retrier", backOffOpts, func(error) bool {
			return false
		})
		myError := errors.New("myError")
		_, err := retrier.Retry(func() (interface{}, error) {
			return nil, myError
		})
		Expect(err).To(Equal(myError))
	})
})

var _ = Describe("RetryWhenErrorTypeMatches", func() {
	var MyErrorType = reflect.TypeOf(&MyError{})

	It("returns true when an error matches an expected type", func() {
		shouldRetryFunc := RetryWhenErrorTypeMatches([]reflect.Type{MyErrorType})
		Expect(shouldRetryFunc(&MyError{})).To(BeTrue())
		Expect(shouldRetryFunc(errors.New("hi"))).To(BeFalse())
	})
})
