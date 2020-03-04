package ghttp

import (
	"bytes"
	"context"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequest(t *testing.T) {
	const (
		validURL   = "https://www.example.com"
		invalidURL = "https://www.example.com^"
	)

	_, err := NewRequest(MethodGet, validURL)
	assert.NoError(t, err)

	_, err = NewRequest(MethodGet, invalidURL)
	assert.Error(t, err)
}

func TestRequest_SetQuery(t *testing.T) {
	params := Params{
		"k1": "hello",
		"k2": []string{"hello", "hi"},
	}

	dummyRequest := &Request{Request: &http.Request{
		URL: &neturl.URL{},
	}}
	dummyRequest.SetQuery(params)
	assert.Equal(t, params.EncodeToURL(true), dummyRequest.URL.RawQuery)
}

func TestRequest_SetHeaders(t *testing.T) {
	headers := Headers{
		"host": "google.com",
		"k1":   "v1",
		"k2":   "v2",
	}
	dummyRequest := &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	dummyRequest.SetHeaders(headers)
	assert.Equal(t, headers["host"], dummyRequest.Host)
	assert.Equal(t, headers["k1"], dummyRequest.Header.Get("K1"))
	assert.Equal(t, headers["k2"], dummyRequest.Header.Get("K2"))
}

func TestRequest_SetUserAgent(t *testing.T) {
	const userAgent = "Go-http-client"

	dummyRequest := &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	dummyRequest.SetUserAgent(userAgent)
	assert.Equal(t, userAgent, dummyRequest.Header.Get("User-Agent"))
}

func TestRequest_SetOrigin(t *testing.T) {
	const origin = "https://www.google.com"

	dummyRequest := &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	dummyRequest.SetOrigin(origin)
	assert.Equal(t, origin, dummyRequest.Header.Get("Origin"))
}

func TestRequest_SetReferer(t *testing.T) {
	const referer = "https://www.google.com"

	dummyRequest := &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	dummyRequest.SetReferer(referer)
	assert.Equal(t, referer, dummyRequest.Header.Get("Referer"))
}

func TestRequest_SetBearerToken(t *testing.T) {
	const token = "ghttp"

	dummyRequest := &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	dummyRequest.SetBearerToken(token)
	assert.Equal(t, "Bearer "+token, dummyRequest.Header.Get("Authorization"))
}

func TestRequest_AddCookies(t *testing.T) {
	dummyRequest := &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	cookies := Cookies{
		"n1": "v1",
		"n2": "v2",
	}
	dummyRequest.AddCookies(cookies)
	assert.Len(t, dummyRequest.Cookies(), len(cookies))
}

func TestRequest_SetBody(t *testing.T) {
	const dummyData = "hello world"

	dummyRequest := &Request{Request: &http.Request{}}
	dummyRequest.SetBody(bytes.NewBuffer([]byte(dummyData)))
	assert.True(t, int64(len(dummyData)) == dummyRequest.ContentLength)
	if assert.NotNil(t, dummyRequest.GetBody) {
		rc, err := dummyRequest.GetBody()
		if assert.NoError(t, err) {
			assert.NotNil(t, rc)
		}
	}

	dummyRequest.SetBody(bytes.NewReader([]byte(dummyData)))
	assert.True(t, int64(len(dummyData)) == dummyRequest.ContentLength)
	if assert.NotNil(t, dummyRequest.GetBody) {
		rc, err := dummyRequest.GetBody()
		if assert.NoError(t, err) {
			assert.NotNil(t, rc)
		}
	}

	dummyRequest.SetBody(strings.NewReader(dummyData))
	assert.True(t, int64(len(dummyData)) == dummyRequest.ContentLength)
	if assert.NotNil(t, dummyRequest.GetBody) {
		rc, err := dummyRequest.GetBody()
		if assert.NoError(t, err) {
			assert.NotNil(t, rc)
		}
	}

	dummyRequest.SetBody(bytes.NewBuffer(nil))
	assert.Zero(t, dummyRequest.ContentLength)
	if assert.NotNil(t, dummyRequest.GetBody) {
		rc, err := dummyRequest.GetBody()
		if assert.NoError(t, err) {
			assert.True(t, rc == http.NoBody)
		}
	}
}

func TestRequest_SetContent(t *testing.T) {
	const dummyData = "hello world"

	dummyRequest := &Request{Request: &http.Request{}}
	dummyRequest.SetContent([]byte(dummyData))
	if assert.NotNil(t, dummyRequest.Body) {
		defer dummyRequest.Body.Close()
		b, err := ioutil.ReadAll(dummyRequest.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, dummyData, string(b))
		}
	}
}

func TestRequest_SetText(t *testing.T) {
	const dummyData = "hello world"

	dummyRequest := &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	dummyRequest.SetText(dummyData)
	assert.Equal(t, "text/plain; charset=utf-8", dummyRequest.Header.Get("Content-Type"))
	if assert.NotNil(t, dummyRequest.Body) {
		defer dummyRequest.Body.Close()
		b, err := ioutil.ReadAll(dummyRequest.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, dummyData, string(b))
		}
	}
}

func TestRequest_SetForm(t *testing.T) {
	form := Form{
		"k1": "hello",
		"k2": []string{"hello", "hi"},
	}
	dummyRequest := &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	dummyRequest.SetForm(form)
	assert.Equal(t, "application/x-www-form-urlencoded", dummyRequest.Header.Get("Content-Type"))
	if assert.NotNil(t, dummyRequest.Body) {
		defer dummyRequest.Body.Close()
		b, err := ioutil.ReadAll(dummyRequest.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, form.EncodeToURL(true), string(b))
		}
	}
}

func TestRequest_SetJSON(t *testing.T) {
	dummyRequest := &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	err := dummyRequest.SetJSON(map[string]interface{}{
		"num": math.Inf(1),
	})
	assert.Error(t, err)

	dummyRequest = &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	err = dummyRequest.SetJSON(map[string]interface{}{
		"msg": "hello world",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, "application/json", dummyRequest.Header.Get("Content-Type"))
		if assert.NotNil(t, dummyRequest.Body) {
			defer dummyRequest.Body.Close()
			b, err := ioutil.ReadAll(dummyRequest.Body)
			if assert.NoError(t, err) {
				assert.Equal(t, `{"msg":"hello world"}`, string(b))
			}
		}
	}
}

func TestRequest_SetFiles(t *testing.T) {
	client := New()

	result := new(postmanResponse)
	resp, err := client.
		Post("https://httpbin.org/post",
			WithFiles(Files{
				"file0": FileFromReader(&dummyBody{s: "hello world", errFlag: errRead}).WithFilename("dummyFile.txt"),
				"file1": MustOpen("./testdata/testfile1.txt"),
				"file2": FileFromReader(strings.NewReader("<p>This is a text file from memory</p>")),
			}),
		)
	require.NoError(t, err)

	err = resp.JSON(result)
	if assert.NoError(t, err) {
		assert.Equal(t, "testfile1.txt", result.Files.GetString("file1"))
		assert.Equal(t, "<p>This is a text file from memory</p>", result.Files.GetString("file2"))
	}
}

func TestRequest_SetContext(t *testing.T) {
	dummyRequest := &Request{Request: &http.Request{
		Header: make(http.Header),
	}}
	assert.Equal(t, context.Background(), dummyRequest.Context())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	dummyRequest.SetContext(ctx)
	assert.Equal(t, ctx, dummyRequest.Context())
}

func TestRequest_EnableRetry(t *testing.T) {
	const (
		dummyData = "hello world"
		n         = 5
	)

	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts%n == 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
		}
	}))
	defer ts.Close()

	client := New()
	resp, err := client.
		Post(ts.URL,
			WithText(dummyData),
			EnableRetry(),
		)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		assert.Equal(t, defaultRetryMaxAttempts, attempts-1)
	}

	resp, err = client.
		Get(ts.URL,
			EnableRetry(
				WithRetryMaxAttempts(n),
			),
		)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, n, attempts)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err = client.
		Get(ts.URL,
			WithContext(ctx),
			EnableRetry(),
		)
	assert.Error(t, err)
}

func TestRequest_EnableClientTrace(t *testing.T) {
	dummyRequest := &Request{}
	assert.False(t, dummyRequest.clientTrace)

	dummyRequest.EnableClientTrace()
	assert.True(t, dummyRequest.clientTrace)
}

func TestRequest_Dump(t *testing.T) {
	req, err := NewRequest(MethodGet, "https://httpbin.org/get")
	if assert.NoError(t, err) {
		_, err = req.Dump(false)
		assert.NoError(t, err)
	}
}
