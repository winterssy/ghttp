package ghttp

import (
	"context"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetrier_ModifyRequest(t *testing.T) {
	const dummyData = "hello world"
	req, _ := NewRequest(MethodPost, "https://httpbin.org/post")
	req.SetBody(&dummyBody{s: dummyData})

	retrier := defaultRetrier()
	err := retrier.modifyRequest(req)
	if assert.NoError(t, err) && assert.NotNil(t, req.GetBody) {
		rc, _ := req.GetBody()
		b, err := ioutil.ReadAll(rc)
		if assert.NoError(t, err) {
			rc.Close()
			assert.Equal(t, dummyData, string(b))
		}
	}

	dummyRequest := &Request{
		Request: &http.Request{
			Method: MethodPost,
			URL: &neturl.URL{
				Scheme: "https",
				Host:   "httpbin.org",
				Path:   "/post",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
			Body:       &dummyBody{errFlag: errRead},
			Host:       "httpbin.org",
		},
	}
	err = retrier.modifyRequest(dummyRequest)
	assert.Equal(t, errAccessDummyBody, err)
}

func TestRetrier_On(t *testing.T) {
	retrier := defaultRetrier()
	dummyResponse := &Response{
		Response: &http.Response{
			StatusCode: http.StatusTooManyRequests,
		},
	}
	assert.True(t, retrier.on(context.Background(), 0, dummyResponse, nil))
	assert.False(t, retrier.on(context.Background(), retrier.maxAttempts, dummyResponse, nil))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.False(t, retrier.on(ctx, retrier.maxAttempts, dummyResponse, nil))

	opt := WithRetryTriggers(func(resp *Response, err error) bool {
		return err != nil || resp.StatusCode != http.StatusOK
	})
	opt(retrier)
	dummyResponse = &Response{
		Response: &http.Response{
			StatusCode: http.StatusInternalServerError,
		},
	}
	assert.True(t, retrier.on(context.Background(), 0, dummyResponse, nil))

	dummyResponse = &Response{
		Response: &http.Response{
			StatusCode: http.StatusOK,
		},
	}
	assert.False(t, retrier.on(context.Background(), 0, dummyResponse, nil))
}

func TestRetryOption(t *testing.T) {
	const maxAttempts = 7
	backoff := NewConstantBackoff(time.Second, true)
	opts := []RetryOption{
		WithRetryMaxAttempts(maxAttempts),
		WithRetryBackoff(backoff),
		WithRetryTriggers(func(resp *Response, err error) bool {
			return err != nil || resp.StatusCode != http.StatusOK
		}),
	}

	retrier := defaultRetrier()
	for _, opt := range opts {
		opt(retrier)
	}
	assert.Equal(t, maxAttempts, retrier.maxAttempts)
	assert.Equal(t, backoff, retrier.backoff)
	assert.NotEmpty(t, retrier.triggers)
}
