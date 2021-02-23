package ghttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func TestResponse_Cookie(t *testing.T) {
	var _cookie = &http.Cookie{
		Name:  "uid",
		Value: "10086",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, _cookie)
	}))
	defer ts.Close()

	client := New()
	resp, err := client.
		Get(ts.URL)
	require.NoError(t, err)

	cookie, err := resp.Cookie(_cookie.Name)
	if assert.NoError(t, err) {
		assert.Equal(t, _cookie.Value, cookie.Value)
	}

	cookie, err = resp.Cookie("uuid")
	if assert.Equal(t, http.ErrNoCookie, err) {
		assert.Nil(t, cookie)
	}
}

func TestResponse_Text(t *testing.T) {
	const dummyData = "你好世界"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var _w io.Writer = w
		if r.Method == MethodGet {
			_w = transform.NewWriter(w, simplifiedchinese.GBK.NewEncoder())
		}
		_w.Write([]byte(dummyData))
	}))
	defer ts.Close()

	client := New()
	resp, err := client.Post(ts.URL)
	require.NoError(t, err)

	data, err := resp.Text()
	if assert.NoError(t, err) {
		assert.Equal(t, dummyData, data)
	}

	resp, err = client.Get(ts.URL)
	require.NoError(t, err)

	data, err = resp.Text(simplifiedchinese.GBK)
	if assert.NoError(t, err) {
		assert.Equal(t, dummyData, data)
	}
}

func TestResponse_H(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"msg":"hello world"}`))
	}))
	defer ts.Close()

	client := New()
	resp, err := client.Get(ts.URL)
	require.NoError(t, err)

	data, err := resp.H()
	if assert.NoError(t, err) {
		assert.Equal(t, "hello world", data.GetString("msg"))
	}
}

func TestResponse_SaveFile(t *testing.T) {
	const testFile = "testdata.txt"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	}))
	defer ts.Close()

	client := New()
	resp, err := client.Get(ts.URL)
	require.NoError(t, err)

	assert.NoError(t, resp.SaveFile(testFile, 0644))
}

func TestResponse_TraceInfo(t *testing.T) {
	client := New()
	resp, _ := client.
		Get("https://httpbin.org/get",
			WithClientTrace(),
		)

	traceInfo := resp.TraceInfo()
	if assert.NotNil(t, traceInfo) {
		assert.True(t, traceInfo.DNSLookupTime >= 0)
		assert.True(t, traceInfo.TCPConnTime >= 0)
		assert.True(t, traceInfo.TLSHandshakeTime >= 0)
		assert.True(t, traceInfo.ConnTime >= 0)
		assert.True(t, traceInfo.ServerTime >= 0)
		assert.True(t, traceInfo.ResponseTime >= 0)
		assert.True(t, traceInfo.TotalTime >= 0)
	}
}

func TestResponse_Dump(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	}))
	defer ts.Close()

	client := New()
	resp, err := client.Get(ts.URL)
	require.NoError(t, err)

	_, err = resp.Dump(true)
	assert.NoError(t, err)
}
