package ghttp

import (
	"context"
	"net/http"
	"time"

	"github.com/winterssy/bufferpool"
)

const (
	defaultRetryMaxAttempts = 3
)

type (
	retrier struct {
		maxAttempts int
		backoff     Backoff
		triggers    []func(resp *Response, err error) bool
	}

	// RetryOption configures a retrier.
	RetryOption func(r *retrier)
)

func defaultRetrier() *retrier {
	return &retrier{
		maxAttempts: defaultRetryMaxAttempts,
		backoff:     NewExponentialBackoff(1*time.Second, 30*time.Second, true),
	}
}

func (r *retrier) modifyRequest(req *Request) (err error) {
	if r.maxAttempts > 0 && req.Body != nil && req.GetBody == nil {
		buf := bufferpool.Get()
		defer buf.Free()
		err = drainBody(req.Body, buf)
		if err == nil {
			req.SetContent(buf.Bytes())
		}
	}
	return
}

// Report whether a request needs a retry.
func (r *retrier) on(ctx context.Context, attemptNum int, resp *Response, err error) bool {
	if ctx.Err() != nil || attemptNum >= r.maxAttempts {
		return false
	}

	if len(r.triggers) == 0 {
		return err != nil || resp.StatusCode == http.StatusTooManyRequests
	}

	for _, trigger := range r.triggers {
		if trigger(resp, err) {
			return true
		}
	}

	return false
}

// WithRetryMaxAttempts is a retry option that specifies the max attempts to a retrier while 0 means no retries.
// By default is 3.
func WithRetryMaxAttempts(n int) RetryOption {
	return func(r *retrier) {
		r.maxAttempts = n
	}
}

// WithRetryBackoff is a retry option that specifies the backoff to a retrier.
// By default is an exponential backoff with jitter whose baseInterval is 1s and maxInterval is 30s.
func WithRetryBackoff(backoff Backoff) RetryOption {
	return func(r *retrier) {
		r.backoff = backoff
	}
}

// WithRetryTriggers is a retry option that specifies the triggers to a retrier
// for determining whether a request needs a retry.
// By default is the error isn't nil or the response status code is 429.
func WithRetryTriggers(triggers ...func(resp *Response, err error) bool) RetryOption {
	return func(r *retrier) {
		r.triggers = triggers
	}
}
