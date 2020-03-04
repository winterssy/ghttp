package ghttp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstantBackoff_Wait(t *testing.T) {
	const (
		interval = 100 * time.Millisecond
	)

	backoff := NewConstantBackoff(interval, false)
	for i := 0; i < 10; i++ {
		assert.Equal(t, interval, backoff.Wait(i, nil, nil))
	}

	backoff = NewConstantBackoff(interval, true)
	for i := 0; i < 10; i++ {
		assert.True(t, backoff.Wait(i, nil, nil) >= interval/2)
		assert.True(t, backoff.Wait(i, nil, nil) <= (interval/2+interval))
	}
}

func TestExponentialBackoff_Wait(t *testing.T) {
	const (
		baseInterval = 1 * time.Second
		maxInterval  = 30 * time.Second
	)

	backoff := NewExponentialBackoff(baseInterval, maxInterval, false)
	for i := 0; i < 10; i++ {
		assert.True(t, backoff.Wait(i, nil, nil) >= baseInterval/2)
		assert.True(t, backoff.Wait(i, nil, nil) <= maxInterval)
	}

	backoff = NewExponentialBackoff(baseInterval, maxInterval, true)
	for i := 0; i < 10; i++ {
		assert.True(t, backoff.Wait(i, nil, nil) >= baseInterval/2)
		assert.True(t, backoff.Wait(i, nil, nil) <= maxInterval)
	}
}

func TestFibonacciBackoff_Wait(t *testing.T) {
	nums := []time.Duration{1, 1, 2, 3, 5, 8, 13, 21, 34, 55}
	backoff := NewFibonacciBackoff(0, time.Second)
	for i, v := range nums {
		assert.True(t, backoff.Wait(i, nil, nil) == v*time.Second)
	}

	nums = []time.Duration{1, 1, 2, 3, 5, 8, 13, 21, 30, 30}
	backoff = NewFibonacciBackoff(30, time.Second)
	for i, v := range nums {
		assert.True(t, backoff.Wait(i, nil, nil) == v*time.Second)
	}
}
