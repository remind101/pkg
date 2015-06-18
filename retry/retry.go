package retry

import (
	"log"
	"reflect"
	"time"

	"github.com/cenkalti/backoff"
)

type BackOffOpts struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxElapsedTime  time.Duration
}

var DefaultBackOffOpts *BackOffOpts = &BackOffOpts{
	InitialInterval: 500 * time.Millisecond,
	MaxInterval:     3 * time.Second,
	MaxElapsedTime:  10 * time.Second}

var RetryOnAnyError = func(error) bool { return true }

type RetryNotifier func(*RetryEvent)

type Retrier struct {
	Name              string
	backOffOpts       *BackOffOpts
	shouldRetryFunc   func(error) bool
	notifyRetryFuncs  []RetryNotifier
	notifyGaveUpFuncs []RetryNotifier
}

func NewRetrier(name string,
	backOffOpts *BackOffOpts, shouldRetryFunc func(error) bool) *Retrier {

	return &Retrier{
		Name: name, backOffOpts: backOffOpts, shouldRetryFunc: shouldRetryFunc,
		notifyRetryFuncs:  []RetryNotifier{logRetry},
		notifyGaveUpFuncs: []RetryNotifier{logGaveUp}}
}

func NewErrorTypeRetrier(name string,
	backOffOpts *BackOffOpts, errorTypes ...interface{}) *Retrier {

	return &Retrier{
		Name:            name,
		backOffOpts:     backOffOpts,
		shouldRetryFunc: RetryWhenErrorTypeMatches(instancesToTypes(errorTypes))}
}

func (r *Retrier) Retry(f func() (interface{}, error)) (interface{}, error) {
	var val interface{}
	var err error
	var next time.Duration

	b := r.newBackOff()
	b.Reset()
	for {
		if val, err = f(); err == nil {
			return val, nil
		}

		if !r.shouldRetryFunc(err) {
			return nil, err
		}

		if next = b.NextBackOff(); next == backoff.Stop {
			r.notifyGaveUp(err)
			return nil, err
		}

		time.Sleep(next)
		r.notifyRetry(err)
	}
}

type RetryEvent struct {
	Retrier *Retrier
	Err     error
}

func (r *Retrier) AddNotifyRetry(f RetryNotifier) {
	r.notifyRetryFuncs = append(r.notifyRetryFuncs, f)
}

func (r *Retrier) AddNotifyGaveUp(f RetryNotifier) {
	r.notifyGaveUpFuncs = append(r.notifyGaveUpFuncs, f)
}

func (r *Retrier) notifyGaveUp(err error) {
	retryEvent := &RetryEvent{Retrier: r, Err: err}
	for _, notifyGaveUpFunc := range r.notifyGaveUpFuncs {
		notifyGaveUpFunc(retryEvent)
	}
}

func (r *Retrier) notifyRetry(err error) {
	retryEvent := &RetryEvent{Retrier: r, Err: err}
	for _, notifyRetryFunc := range r.notifyRetryFuncs {
		notifyRetryFunc(retryEvent)
	}
}

func logRetry(re *RetryEvent) {
	log.Printf("Retrying error=%s count#retry.%s.retry_count=1\n",
		re.Err.Error(), re.Retrier.Name)
}

func logGaveUp(re *RetryEvent) {
	log.Printf("Retrying error=%s count#retry.%s.gave_up_count=1\n",
		re.Err.Error(), re.Retrier.Name)
}

func RetryWhenErrorTypeMatches(errorTypes []reflect.Type) func(error) bool {
	errorTypeSet := make(map[reflect.Type]bool)
	for _, t := range errorTypes {
		errorTypeSet[t] = true
	}
	return func(e error) bool {
		errorType := reflect.TypeOf(e)
		return errorTypeSet[errorType] == true
	}
}

func (r *Retrier) SetBackOffOpts(b *BackOffOpts) {
	r.backOffOpts = b
}

func (r *Retrier) newBackOff() backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = r.backOffOpts.InitialInterval
	b.MaxInterval = r.backOffOpts.MaxInterval
	b.MaxElapsedTime = r.backOffOpts.MaxElapsedTime
	return b
}

func instancesToTypes(instances []interface{}) []reflect.Type {
	types := []reflect.Type{}
	for _, instance := range instances {
		types = append(types, reflect.TypeOf(instance))
	}
	return types
}
