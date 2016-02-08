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

	numTries := 0
	b := r.newBackOff()
	b.Reset()
	for {
		numTries++
		if val, err = f(); err == nil {
			return val, nil
		}

		if !r.shouldRetryFunc(err) {
			return val, err
		}

		if next = b.NextBackOff(); next == backoff.Stop {
			r.notifyGaveUp(err, numTries)
			return val, err
		}

		time.Sleep(next)
		r.notifyRetry(err, numTries)
	}
}

type RetryEvent struct {
	Retrier *Retrier
	Err     error
	NumTries int
}

func (r *Retrier) AddNotifyRetry(f RetryNotifier) {
	r.notifyRetryFuncs = append(r.notifyRetryFuncs, f)
}

func (r *Retrier) AddNotifyGaveUp(f RetryNotifier) {
	r.notifyGaveUpFuncs = append(r.notifyGaveUpFuncs, f)
}

func (r *Retrier) notifyGaveUp(err error, numTries int) {
	retryEvent := &RetryEvent{Retrier: r, Err: err, NumTries: numTries}
	for _, notifyGaveUpFunc := range r.notifyGaveUpFuncs {
		notifyGaveUpFunc(retryEvent)
	}
}

func (r *Retrier) notifyRetry(err error, numTries int) {
	retryEvent := &RetryEvent{Retrier: r, Err: err, NumTries: numTries}
	for _, notifyRetryFunc := range r.notifyRetryFuncs {
		notifyRetryFunc(retryEvent)
	}
}

func logRetry(re *RetryEvent) {
	log.Printf("Retrying after %d tries: error=%s count#retry.%s.retry_count=1\n",
		re.NumTries, re.Err.Error(), re.Retrier.Name)
}

func logGaveUp(re *RetryEvent) {
	log.Printf("Giving up after %d tries: error=%s count#retry.%s.gave_up_count=1\n",
		re.NumTries, re.Err.Error(), re.Retrier.Name)
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
