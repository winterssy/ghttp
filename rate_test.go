package ghttp

import (
	"context"
	"net/http"
	neturl "net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimiter_Enter(t *testing.T) {
	dummyRequest := &Request{Request: &http.Request{
		URL: &neturl.URL{
			Scheme: "https",
			Host:   "httpbin.org",
			Path:   "/get",
		},
		Header: make(http.Header),
	}}

	rl := &rateLimiter{base: rate.NewLimiter(1, 10)}
	err := rl.Enter(dummyRequest)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	dummyRequest.SetContext(ctx)
	cancel()
	err = rl.Enter(dummyRequest)
	assert.Equal(t, ctx.Err(), err)
}

func TestConcurrency_Enter(t *testing.T) {
	dummyRequest := &Request{Request: &http.Request{
		URL: &neturl.URL{
			Scheme: "https",
			Host:   "httpbin.org",
			Path:   "/get",
		},
		Header: make(http.Header),
	}}

	c := &concurrency{ch: make(chan struct{}, 1)}
	err := c.Enter(dummyRequest)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	dummyRequest.SetContext(ctx)
	cancel()
	err = c.Enter(dummyRequest)
	assert.Equal(t, ctx.Err(), err)
}
