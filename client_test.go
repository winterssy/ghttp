package ghttp

import (
	"compress/gzip"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestClient_SetProxy(t *testing.T) {
	const (
		proxyURL = "http://127.0.0.1:1081"
	)

	client := New()
	client.SetProxy(ProxyURL(proxyURL))
	transport := client.Transport.(*http.Transport)
	require.NotNil(t, transport.Proxy)

	req, _ := http.NewRequest("GET", "https://www.google.com", nil)
	fixedURL, err := transport.Proxy(req)
	if assert.NoError(t, err) {
		assert.Equal(t, proxyURL, fixedURL.String())
	}

	client.SetProxy(nil)
	assert.Nil(t, transport.Proxy)
}

func TestClient_SetTLSClientConfig(t *testing.T) {
	config := &tls.Config{}
	client := New()
	client.SetTLSClientConfig(config)
	transport := client.Transport.(*http.Transport)
	assert.NotNil(t, transport.TLSClientConfig)
}

func TestClient_AddClientCerts(t *testing.T) {
	cert := tls.Certificate{}
	client := New()
	client.AddClientCerts(cert)
	transport := client.Transport.(*http.Transport)
	if assert.NotNil(t, transport.TLSClientConfig) {
		assert.Len(t, transport.TLSClientConfig.Certificates, 1)
	}
}

func TestClient_AddRootCerts(t *testing.T) {
	const (
		pemFile = "./testdata/root-ca.pem"
	)

	pemCerts, err := ioutil.ReadFile(pemFile)
	require.NoError(t, err)

	client := New()
	assert.True(t, client.AddRootCerts(pemCerts))
}

func TestClient_DisableTLSVerify(t *testing.T) {
	client := New()
	client.DisableTLSVerify()
	transport := client.Transport.(*http.Transport)
	if assert.NotNil(t, transport.TLSClientConfig) {
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	}
}

func TestClient_AddCookies(t *testing.T) {
	const (
		validURL   = "https://api.example.com"
		invalidURL = "https://api.example.com^"
	)

	var (
		_cookie = &http.Cookie{
			Name:  "uid",
			Value: "10086",
		}
	)

	client := New()
	client.AddCookies(invalidURL, _cookie)
	assert.Empty(t, client.Cookies(invalidURL))

	client.AddCookies(validURL, _cookie)
	cookie, err := client.Cookie(validURL, _cookie.Name)
	if assert.NoError(t, err) {
		assert.Equal(t, _cookie.Value, cookie.Value)
	}

	cookie, err = client.Cookie(validURL, "uuid")
	if assert.Equal(t, http.ErrNoCookie, err) {
		assert.Nil(t, cookie)
	}
}

func TestClient_EnableRateLimiting(t *testing.T) {
	const (
		r           rate.Limit = 1
		bursts                 = 5
		concurrency            = 10
	)

	var counter uint64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&counter, 1)
	}))
	defer ts.Close()

	client := New()
	client.EnableRateLimiting(rate.NewLimiter(r, bursts))

	wg := new(sync.WaitGroup)
	start := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			_, _ = client.Get(ts.URL)
			wg.Done()
		}()
	}
	wg.Wait()
	end := time.Since(start)

	if assert.Equal(t, uint64(concurrency), atomic.LoadUint64(&counter)) {
		assert.True(t, end >= ((concurrency-bursts)*time.Second))
	}
}

func TestClient_SetMaxConcurrency(t *testing.T) {
	const (
		n           = 4
		concurrency = 20
		sleep       = time.Second
	)

	var counter uint64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(sleep)
		atomic.AddUint64(&counter, 1)
	}))
	defer ts.Close()

	client := New()
	client.SetMaxConcurrency(n)

	wg := new(sync.WaitGroup)
	start := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			_, _ = client.Get(ts.URL)
			wg.Done()
		}()
	}
	wg.Wait()
	end := time.Since(start)

	if assert.Equal(t, uint64(concurrency), atomic.LoadUint64(&counter)) {
		assert.True(t, end >= (concurrency/n)*sleep)
		assert.True(t, end <= concurrency*sleep)
	}
}

func TestClient_EnableDebugging(t *testing.T) {
	client := New()
	client.EnableDebugging(ioutil.Discard, true)

	resp, err := client.Post("https://httpbin.org/post",
		WithForm(Form{
			"k1": "v1",
			"k2": "v2",
		}),
	)
	require.NoError(t, err)

	result := new(postmanResponse)
	err = resp.JSON(result)
	if assert.NoError(t, err) {
		assert.Equal(t, "v1", result.Form.GetString("k1"))
		assert.Equal(t, "v2", result.Form.GetString("k2"))
	}
}

func TestClient_AutoGzip(t *testing.T) {
	const dummyData = "hello world"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Encoding", "gzip")
		zw := gzip.NewWriter(w)
		_, _ = zw.Write([]byte(dummyData))
		zw.Close()
	}))
	defer ts.Close()

	client := New()
	resp, err := client.
		Get(ts.URL,
			WithHeaders(Headers{
				"Accept-Encoding": "gzip",
			}),
		)
	if assert.NoError(t, err) {
		data, err := resp.Text()
		if assert.NoError(t, err) {
			assert.Equal(t, dummyData, data)
		}
	}
}

func TestClient_Get(t *testing.T) {
	client := New()
	resp, err := client.Get("https://httpbin.org/get")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestClient_Head(t *testing.T) {
	client := New()
	resp, err := client.Head("https://httpbin.org")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestClient_Post(t *testing.T) {
	client := New()
	resp, err := client.Post("https://httpbin.org/post")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestClient_Put(t *testing.T) {
	client := New()
	resp, err := client.Put("https://httpbin.org/put")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestClient_Patch(t *testing.T) {
	client := New()
	resp, err := client.Patch("https://httpbin.org/patch")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestClient_Delete(t *testing.T) {
	client := New()
	resp, err := client.Delete("https://httpbin.org/delete")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestClient_Options(t *testing.T) {
	client := New()
	resp, err := client.Options("https://httpbin.org")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestClient_Send(t *testing.T) {
	client := New()
	_, err := client.Send(MethodPost, "https://httpbin.org/post", func(req *Request) error {
		return errAccessDummyBody
	})
	assert.Equal(t, errAccessDummyBody, err)
}

func TestClient_Do(t *testing.T) {
	client := New()
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
	dummyRequest.EnableRetrier()
	_, err := client.Do(dummyRequest)
	assert.Equal(t, errAccessDummyBody, err)

	var dummyRequestHook RequestHook = func(req *Request) error {
		return errAccessDummyBody
	}
	client.RegisterBeforeRequestCallbacks(dummyRequestHook)
	_, err = client.Do(dummyRequest)
	assert.Equal(t, errAccessDummyBody, err)
}
