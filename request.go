package ghttp

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/winterssy/gjson"
)

// Common HTTP methods.
const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH"
	MethodDelete  = "DELETE"
	MethodOptions = "OPTIONS"
	MethodConnect = "CONNECT"
	MethodTrace   = "TRACE"
)

type (
	// Request is a wrapper around an http.Request.
	Request struct {
		*http.Request
		retrier     *retrier
		clientTrace bool
	}

	// RequestHook is a function that implements BeforeRequestCallback interface.
	// It provides a elegant way to configure a Request.
	RequestHook func(req *Request) error
)

// Enter implements BeforeRequestCallback interface.
func (rh RequestHook) Enter(req *Request) error {
	return rh(req)
}

// NewRequest returns a new Request given a method, URL.
func NewRequest(method string, url string) (*Request, error) {
	rawRequest, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	return &Request{
		Request: rawRequest,
	}, nil
}

// SetQuery sets query parameters for req.
// It replaces any existing values.
func (req *Request) SetQuery(params Params) {
	query := req.URL.Query()
	for k, vs := range params.Decode() {
		query[k] = vs
	}
	req.URL.RawQuery = query.Encode()
}

// SetHeaders sets headers for req.
// It replaces any existing values.
func (req *Request) SetHeaders(headers Headers) {
	for k, vs := range headers.Decode() {
		k = http.CanonicalHeaderKey(k)
		if k == "Host" && len(vs) > 0 {
			req.Host = vs[0]
		} else {
			req.Header[k] = vs
		}
	}
}

// SetContentType sets Content-Type header value for req.
func (req *Request) SetContentType(contentType string) {
	req.Header.Set("Content-Type", contentType)
}

// SetUserAgent sets User-Agent header value for req.
func (req *Request) SetUserAgent(userAgent string) {
	req.Header.Set("User-Agent", userAgent)
}

// SetOrigin sets Origin header value for req.
func (req *Request) SetOrigin(origin string) {
	req.Header.Set("Origin", origin)
}

// SetReferer sets Referer header value for req.
func (req *Request) SetReferer(referer string) {
	req.Header.Set("Referer", referer)
}

// SetBearerToken sets bearer token for req.
func (req *Request) SetBearerToken(token string) {
	req.Header.Set("Authorization", "Bearer "+token)
}

// AddCookies adds cookies to req.
func (req *Request) AddCookies(cookies Cookies) {
	for _, c := range cookies.Decode() {
		req.AddCookie(c)
	}
}

// SetBody sets body for req.
func (req *Request) SetBody(body io.Reader) {
	req.Body = toReadCloser(body)
	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			req.ContentLength = int64(v.Len())
			buf := v.Bytes()
			req.GetBody = func() (io.ReadCloser, error) {
				r := bytes.NewReader(buf)
				return ioutil.NopCloser(r), nil
			}
		case *bytes.Reader:
			req.ContentLength = int64(v.Len())
			snapshot := *v
			req.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		case *strings.Reader:
			req.ContentLength = int64(v.Len())
			snapshot := *v
			req.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		default:
			// This is where we'd set it to -1 (at least
			// if body != NoBody) to mean unknown, but
			// that broke people during the Go 1.8 testing
			// period. People depend on it being 0 I
			// guess. Maybe retry later. See Issue 18117.
		}
		// For client requests, Request.ContentLength of 0
		// means either actually 0, or unknown. The only way
		// to explicitly say that the ContentLength is zero is
		// to set the Body to nil. But turns out too much code
		// depends on NewRequest returning a non-nil Body,
		// so we use a well-known ReadCloser variable instead
		// and have the http package also treat that sentinel
		// variable to mean explicitly zero.
		if req.GetBody != nil && req.ContentLength == 0 {
			req.Body = http.NoBody
			req.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
		}
	}
}

// SetContent sets bytes payload for req.
func (req *Request) SetContent(content []byte) {
	req.SetBody(bytes.NewReader(content))
}

// SetText sets plain text payload for req.
func (req *Request) SetText(text string) {
	req.SetBody(strings.NewReader(text))
	req.SetContentType("text/plain; charset=utf-8")
}

// SetForm sets form payload for req.
func (req *Request) SetForm(form Form) {
	req.SetBody(strings.NewReader(form.EncodeToURL(true)))
	req.SetContentType("application/x-www-form-urlencoded")
}

// SetJSON sets JSON payload for req.
func (req *Request) SetJSON(data interface{}, opts ...func(enc *gjson.Encoder)) error {
	b, err := gjson.Encode(data, opts...)
	if err != nil {
		return err
	}

	req.SetBody(bytes.NewReader(b))
	req.SetContentType("application/json")
	return nil
}

// SetFiles sets files payload for req.
func (req *Request) SetFiles(files Files) {
	formData := NewMultipart(files)
	req.SetBody(formData)
	req.SetContentType(formData.ContentType())
}

// SetContext sets context for req.
func (req *Request) SetContext(ctx context.Context) {
	req.Request = req.WithContext(ctx)
}

// EnableRetry enables retry for req.
func (req *Request) EnableRetry(opts ...RetryOption) {
	retrier := defaultRetrier()
	for _, opt := range opts {
		opt(retrier)
	}
	req.retrier = retrier
}

// EnableClientTrace enables client trace for req using httptrace.ClientTrace.
func (req *Request) EnableClientTrace() {
	req.clientTrace = true
}

// Dump returns the HTTP/1.x wire representation of req.
func (req *Request) Dump(withBody bool) ([]byte, error) {
	return httputil.DumpRequestOut(req.Request, withBody)
}

// WithQuery is a request hook to set query parameters.
// It replaces any existing values.
func WithQuery(params Params) RequestHook {
	return func(req *Request) error {
		req.SetQuery(params)
		return nil
	}
}

// WithHeaders is a request hook to set headers.
// It replaces any existing values.
func WithHeaders(headers Headers) RequestHook {
	return func(req *Request) error {
		req.SetHeaders(headers)
		return nil
	}
}

// WithContentType is a request hook to set Content-Type header value.
func WithContentType(contentType string) RequestHook {
	return func(req *Request) error {
		req.SetContentType(contentType)
		return nil
	}
}

// WithUserAgent is a request hook to set User-Agent header value.
func WithUserAgent(userAgent string) RequestHook {
	return func(req *Request) error {
		req.SetUserAgent(userAgent)
		return nil
	}
}

// WithOrigin is a request hook to set Origin header value.
func WithOrigin(origin string) RequestHook {
	return func(req *Request) error {
		req.SetOrigin(origin)
		return nil
	}
}

// WithReferer is a request hook to set Referer header value.
func WithReferer(referer string) RequestHook {
	return func(req *Request) error {
		req.SetReferer(referer)
		return nil
	}
}

// WithBasicAuth is a request hook to set basic authentication.
func WithBasicAuth(username string, password string) RequestHook {
	return func(req *Request) error {
		req.SetBasicAuth(username, password)
		return nil
	}
}

// WithBearerToken is a request hook to set bearer token.
func WithBearerToken(token string) RequestHook {
	return func(req *Request) error {
		req.SetBearerToken(token)
		return nil
	}
}

// WithCookies is a request hook to add cookies.
func WithCookies(cookies Cookies) RequestHook {
	return func(req *Request) error {
		req.AddCookies(cookies)
		return nil
	}
}

// WithBody is a request hook to set body.
func WithBody(body io.Reader) RequestHook {
	return func(req *Request) error {
		req.SetBody(body)
		return nil
	}
}

// WithContent is a request hook to set bytes payload.
func WithContent(content []byte) RequestHook {
	return func(req *Request) error {
		req.SetContent(content)
		return nil
	}
}

// WithText is a request hook to set plain text payload.
func WithText(text string) RequestHook {
	return func(req *Request) error {
		req.SetText(text)
		return nil
	}
}

// WithForm is a request hook to set form payload.
func WithForm(form Form) RequestHook {
	return func(req *Request) error {
		req.SetForm(form)
		return nil
	}
}

// WithJSON is a request hook to set JSON payload.
func WithJSON(data interface{}, opts ...func(enc *gjson.Encoder)) RequestHook {
	return func(req *Request) error {
		return req.SetJSON(data, opts...)
	}
}

// WithFiles is a request hook to set files payload.
func WithFiles(files Files) RequestHook {
	return func(req *Request) error {
		req.SetFiles(files)
		return nil
	}
}

// WithContext is a request hook to set context.
func WithContext(ctx context.Context) RequestHook {
	return func(req *Request) error {
		req.SetContext(ctx)
		return nil
	}
}

// EnableRetry is a request hook to enable retry.
func EnableRetry(opts ...RetryOption) RequestHook {
	return func(req *Request) error {
		req.EnableRetry(opts...)
		return nil
	}
}

// EnableClientTrace is a request hook to enable client trace.
func EnableClientTrace() RequestHook {
	return func(req *Request) error {
		req.EnableClientTrace()
		return nil
	}
}
