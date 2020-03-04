package ghttp

import (
	"math"
	"math/rand"
	"time"
)

type (
	// Backoff is the interface defines a backoff for a retrier. It is called
	// after a failing request to determine the amount of time
	// that should pass before trying again.
	Backoff interface {
		// Wait returns the duration to wait before retrying a request.
		Wait(attemptNum int, resp *Response, err error) time.Duration
	}

	constantBackoff struct {
		interval time.Duration
		jitter   bool
	}

	exponentialBackoff struct {
		baseInterval time.Duration
		maxInterval  time.Duration
		jitter       bool
	}

	fibonacciBackoff struct {
		maxValue int
		interval time.Duration
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewConstantBackoff provides a callback for the retry policy which
// will perform constant backoff with jitter based on interval.
func NewConstantBackoff(interval time.Duration, jitter bool) Backoff {
	return &constantBackoff{
		interval: interval,
		jitter:   jitter,
	}
}

// Wait implements Backoff interface.
func (cb *constantBackoff) Wait(int, *Response, error) time.Duration {
	if !cb.jitter {
		return cb.interval
	}

	return cb.interval/2 + time.Duration(rand.Int63n(int64(cb.interval)))
}

// NewExponentialBackoff provides a callback for the retry policy which
// will perform exponential backoff with jitter based on the attempt number and limited
// by baseInterval and maxInterval.
// See: https://aws.amazon.com/cn/blogs/architecture/exponential-backoff-and-jitter/
func NewExponentialBackoff(baseInterval, maxInterval time.Duration, jitter bool) Backoff {
	return &exponentialBackoff{
		baseInterval: baseInterval,
		maxInterval:  maxInterval,
		jitter:       jitter,
	}
}

// Wait implements Backoff interface.
func (eb *exponentialBackoff) Wait(attemptNum int, _ *Response, _ error) time.Duration {
	temp := math.Min(float64(eb.maxInterval), float64(eb.baseInterval)*math.Exp2(float64(attemptNum)))
	if !eb.jitter {
		return time.Duration(temp)
	}

	n := int64(temp / 2)
	return time.Duration(n + rand.Int63n(n))
}

// NewFibonacciBackoff provides a callback for the retry policy which
// will perform fibonacci backoff based on the attempt number and limited by maxValue.
// If maxValue less than or equal to zero, it means no limit.
func NewFibonacciBackoff(maxValue int, interval time.Duration) Backoff {
	return &fibonacciBackoff{
		maxValue: maxValue,
		interval: interval,
	}
}

// Wait implements Backoff interface.
func (fb *fibonacciBackoff) Wait(attemptNum int, _ *Response, _ error) time.Duration {
	a, b := 0, 1
	for ; attemptNum >= 0; attemptNum-- {
		if fb.maxValue > 0 && b >= fb.maxValue {
			a = fb.maxValue
			break
		}
		a, b = b, a+b
	}
	return time.Duration(a) * fb.interval
}
