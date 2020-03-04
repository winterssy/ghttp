package ghttp

import (
	"net/http"
	neturl "net/url"
	"sort"
	"strings"

	"github.com/winterssy/gjson"
)

type (
	// KV maps a string key to an interface{} type value,
	// It's typically used for request query parameters, form data or headers.
	KV map[string]interface{}

	// Params is an alias of KV, used for for request query parameters.
	Params = KV

	// Form is an alias of KV, used for request form data.
	Form = KV

	// Headers is an alias of KV, used for request headers.
	Headers = KV

	// Cookies is a shortcut for map[string]string, used for request cookies.
	Cookies map[string]string

	// BeforeRequestCallback is the interface that defines the manipulations before sending a request.
	BeforeRequestCallback interface {
		// Enter is called when a request is about to begin.
		// If a non-nil error is returned, ghttp will cancel the request.
		Enter(req *Request) error
	}

	// AfterResponseCallback is the interface that defines the manipulations after receiving a response.
	AfterResponseCallback interface {
		// Exit is called when a request ends.
		Exit(resp *Response, err error)
	}
)

// Decode translates kv and returns the equivalent request query parameters, form data or headers.
// It ignores any unexpected key-value pair.
func (kv KV) Decode() map[string][]string {
	vv := make(map[string][]string, len(kv))
	for k, v := range kv {
		if vs := toStrings(v); len(vs) > 0 {
			vv[k] = vs
		}
	}
	return vv
}

// EncodeToURL encodes kv into URL form sorted by key if kv is considered as request query parameters or form data.
func (kv KV) EncodeToURL(escape bool) string {
	vv := kv.Decode()
	keys := make([]string, 0, len(vv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		vs := vv[k]
		for _, v := range vs {
			if sb.Len() > 0 {
				sb.WriteByte('&')
			}

			if escape {
				k = neturl.QueryEscape(k)
				v = neturl.QueryEscape(v)
			}

			sb.WriteString(k)
			sb.WriteByte('=')
			sb.WriteString(v)
		}
	}
	return sb.String()
}

// EncodeToJSON returns the JSON encoding of kv.
// If there is an error, it returns "{}".
func (kv KV) EncodeToJSON(opts ...func(enc *gjson.Encoder)) string {
	s, err := gjson.EncodeToString(kv, opts...)
	if err != nil {
		return "{}"
	}

	return s
}

// Decode translates c and returns the equivalent request cookies.
func (c Cookies) Decode() []*http.Cookie {
	cookies := make([]*http.Cookie, 0, len(c))
	for k, v := range c {
		cookies = append(cookies, &http.Cookie{
			Name:  k,
			Value: v,
		})
	}
	return cookies
}
