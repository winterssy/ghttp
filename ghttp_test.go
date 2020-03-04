package ghttp

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/winterssy/gjson"
)

type (
	postmanResponse struct {
		Args          H      `json:"args,omitempty"`
		Authenticated bool   `json:"authenticated,omitempty"`
		Cookies       H      `json:"cookies,omitempty"`
		Data          string `json:"data,omitempty"`
		Files         H      `json:"files,omitempty"`
		Form          H      `json:"form,omitempty"`
		Headers       H      `json:"headers,omitempty"`
		JSON          H      `json:"json,omitempty"`
		Method        string `json:"method,omitempty"`
		Origin        string `json:"origin,omitempty"`
		Token         string `json:"token,omitempty"`
		URL           string `json:"url,omitempty"`
		User          string `json:"user,omitempty"`
	}
)

func TestKV_Decode(t *testing.T) {
	var (
		v1 = "hello world"
		v2 = []string{"hello", "world"}
	)
	v := KV{
		"k1":      v1,
		"k2":      v2,
		"invalid": 1 + 2i,
	}
	vv := v.Decode()
	assert.Len(t, vv, 2)
	assert.Equal(t, []string{v1}, vv["k1"])
	assert.Equal(t, v2, vv["k2"])
}

func TestKV_EncodeToURL(t *testing.T) {
	var (
		v1 = "hello"
		v2 = []string{"hello", "hi"}
	)
	v := KV{
		"expr": "1+2",
		"k1":   v1,
		"k2":   v2,
	}
	want := "expr=1%2B2&k1=hello&k2=hello&k2=hi"
	assert.Equal(t, want, v.EncodeToURL(true))

	want = "expr=1+2&k1=hello&k2=hello&k2=hi"
	assert.Equal(t, want, v.EncodeToURL(false))
}

func TestKV_EncodeToJSON(t *testing.T) {
	v := KV{
		"text": "<p>Hello World</p>",
	}
	want := "{\"text\":\"<p>Hello World</p>\"}"
	assert.Equal(t, want, v.EncodeToJSON(func(enc *gjson.Encoder) {
		enc.SetEscapeHTML(false)
	}))

	v = KV{
		"num": math.Inf(1),
	}
	assert.Equal(t, "{}", v.EncodeToJSON())
}

func TestCookies_Decode(t *testing.T) {
	c := Cookies{
		"n1": "v1",
		"n2": "v2",
	}
	assert.Len(t, c.Decode(), 2)
}
