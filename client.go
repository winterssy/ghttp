package ghttp

import (
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	neturl "net/url"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
	"golang.org/x/time/rate"
)

const (
	defaultTimeout = 120 * time.Second
)

var (
	// DefaultClient is a global Client used by the global methods such as Get, Post, etc.
	DefaultClient = New()

	// ProxyFromEnvironment is an alias of http.ProxyFromEnvironment.
	ProxyFromEnvironment = http.ProxyFromEnvironment
)

// HTTP verbs via the global Client.
var (
	Get     = DefaultClient.Get
	Head    = DefaultClient.Head
	Post    = DefaultClient.Post
	Put     = DefaultClient.Put
	Patch   = DefaultClient.Patch
	Delete  = DefaultClient.Delete
	Options = DefaultClient.Options
	Send    = DefaultClient.Send
)

type (
	// Client is a wrapper around an http.Client.
	Client struct {
		*http.Client
		beforeRequestCallbacks []BeforeRequestCallback
		afterResponseCallbacks []AfterResponseCallback
	}
)

// New returns a new Client.
func New() *Client {
	jar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	client := &http.Client{
		Transport: DefaultTransport(),
		Jar:       jar,
		Timeout:   defaultTimeout,
	}
	return &Client{
		Client: client,
	}
}

// NoRedirect is a redirect policy that makes the Client not follow redirects.
func NoRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

// MaxRedirects returns a redirect policy for limiting maximum redirects followed up by n.
func MaxRedirects(n int) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= n {
			return http.ErrUseLastResponse
		}

		return nil
	}
}

// ProxyURL returns a proxy function (for use in an http.Transport) given a URL.
func ProxyURL(url string) func(*http.Request) (*neturl.URL, error) {
	return func(*http.Request) (*neturl.URL, error) {
		return neturl.Parse(url)
	}
}

// SetProxy specifies a function to return a proxy for a given Request while nil indicates no proxy.
// By default is ProxyFromEnvironment.
func (c *Client) SetProxy(proxy func(*http.Request) (*neturl.URL, error)) {
	c.Transport.(*http.Transport).Proxy = proxy
}

// SetTLSClientConfig specifies the TLS configuration to use with tls.Client.
func (c *Client) SetTLSClientConfig(config *tls.Config) {
	c.Transport.(*http.Transport).TLSClientConfig = config
}

// AddClientCerts adds client certificates to c.
func (c *Client) AddClientCerts(certs ...tls.Certificate) {
	t := c.Transport.(*http.Transport)
	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}
	t.TLSClientConfig.Certificates = append(t.TLSClientConfig.Certificates, certs...)
}

// AddRootCerts attempts to add root certificates to c from a series of PEM encoded certificates
// and reports whether any certificates were successfully added.
func (c *Client) AddRootCerts(pemCerts []byte) bool {
	t := c.Transport.(*http.Transport)
	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}
	if t.TLSClientConfig.RootCAs == nil {
		t.TLSClientConfig.RootCAs = x509.NewCertPool()
	}
	return t.TLSClientConfig.RootCAs.AppendCertsFromPEM(pemCerts)
}

// DisableTLSVerify makes c not verify the server's TLS certificate.
func (c *Client) DisableTLSVerify() {
	t := c.Transport.(*http.Transport)
	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}
	t.TLSClientConfig.InsecureSkipVerify = true
}

// AddCookies adds cookies to send in a request for the given URL to cookie jar.
func (c *Client) AddCookies(url string, cookies ...*http.Cookie) {
	u, err := neturl.Parse(url)
	if err == nil {
		c.Jar.SetCookies(u, cookies)
	}
}

// Cookies returns the cookies to send in a request for the given URL from cookie jar.
func (c *Client) Cookies(url string) (cookies []*http.Cookie) {
	u, err := neturl.Parse(url)
	if err == nil {
		cookies = c.Jar.Cookies(u)
	}
	return
}

// Cookie returns the named cookie to send in a request for the given URL from cookie jar.
// If multiple cookies match the given name, only one cookie will be returned.
func (c *Client) Cookie(url string, name string) (*http.Cookie, error) {
	return findCookie(name, c.Cookies(url))
}

// RegisterBeforeRequestCallbacks appends c's before request callbacks.
func (c *Client) RegisterBeforeRequestCallbacks(callbacks ...BeforeRequestCallback) {
	c.beforeRequestCallbacks = append(c.beforeRequestCallbacks, callbacks...)
}

// RegisterAfterResponseCallbacks appends c's after response callbacks.
func (c *Client) RegisterAfterResponseCallbacks(callbacks ...AfterResponseCallback) {
	c.afterResponseCallbacks = append(c.afterResponseCallbacks, callbacks...)
}

// EnableRateLimiting adds a callback to c for limiting outbound requests
// given a rate.Limiter (provided by golang.org/x/time/rate package).
func (c *Client) EnableRateLimiting(limiter *rate.Limiter) {
	c.RegisterBeforeRequestCallbacks(&rateLimiter{base: limiter})
}

// SetMaxConcurrency adds a callback to c for limiting the concurrent outbound requests up by n.
func (c *Client) SetMaxConcurrency(n int) {
	callback := &concurrency{ch: make(chan struct{}, n)}
	c.RegisterBeforeRequestCallbacks(callback)
	c.RegisterAfterResponseCallbacks(callback)
}

// EnableDebugging adds a callback to c for debugging.
// ghttp will dump the request and response details to w, like "curl -v".
func (c *Client) EnableDebugging(w io.Writer, body bool) {
	callback := &debugger{out: w, body: body}
	c.RegisterBeforeRequestCallbacks(callback)
	c.RegisterAfterResponseCallbacks(callback)
}

// Get makes a GET HTTP request.
func (c *Client) Get(url string, hooks ...RequestHook) (*Response, error) {
	return c.Send(MethodGet, url, hooks...)
}

// Head makes a HEAD HTTP request.
func (c *Client) Head(url string, hooks ...RequestHook) (*Response, error) {
	return c.Send(MethodHead, url, hooks...)
}

// Post makes a POST HTTP request.
func (c *Client) Post(url string, hooks ...RequestHook) (*Response, error) {
	return c.Send(MethodPost, url, hooks...)
}

// Put makes a PUT HTTP request.
func (c *Client) Put(url string, hooks ...RequestHook) (*Response, error) {
	return DefaultClient.Send(MethodPut, url, hooks...)
}

// Patch makes a PATCH HTTP request.
func (c *Client) Patch(url string, hooks ...RequestHook) (*Response, error) {
	return c.Send(MethodPatch, url, hooks...)
}

// Delete makes a DELETE HTTP request.
func (c *Client) Delete(url string, hooks ...RequestHook) (*Response, error) {
	return c.Send(MethodDelete, url, hooks...)
}

// Options makes a OPTIONS HTTP request.
func (c *Client) Options(url string, hooks ...RequestHook) (*Response, error) {
	return c.Send(MethodOptions, url, hooks...)
}

// Send makes an HTTP request using a particular method.
func (c *Client) Send(method string, url string, hooks ...RequestHook) (*Response, error) {
	req, err := NewRequest(method, url)
	if err == nil {
		for _, hook := range hooks {
			if err = hook(req); err != nil {
				break
			}
		}
	}
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

// Do sends a request and returns its response.
func (c *Client) Do(req *Request) (resp *Response, err error) {
	if err = c.onBeforeRequest(req); err != nil {
		return
	}

	if req.retrier != nil {
		if err = req.retrier.modifyRequest(req); err != nil {
			return
		}
	}

	resp, err = c.doWithRetry(req)
	c.onAfterResponse(resp, err)
	return
}

func (c *Client) onBeforeRequest(req *Request) (err error) {
	for _, callback := range c.beforeRequestCallbacks {
		if err = callback.Enter(req); err != nil {
			break
		}
	}
	return
}

func (c *Client) doWithRetry(req *Request) (*Response, error) {
	var err error
	var sleep time.Duration
	resp := new(Response)
	for attemptNum := 0; ; attemptNum++ {
		if req.clientTrace {
			ct := &clientTrace{start: time.Now()}
			ct.modifyRequest(req)
			resp.clientTrace = ct
		}
		resp.Response, err = c.do(req.Request)
		if req.clientTrace {
			resp.clientTrace.done()
		}

		if req.retrier == nil || !req.retrier.on(req.Context(), attemptNum, resp, err) {
			c.CloseIdleConnections()
			return resp, err
		}

		sleep = req.retrier.backoff.Wait(attemptNum, resp, err)
		// Drain Response.Body to enable TCP/TLS connection reuse
		if err == nil && drainBody(resp.Body, ioutil.Discard) != http.ErrBodyReadAfterClose {
			resp.Body.Close()
		}

		if req.GetBody != nil {
			req.Body, _ = req.GetBody()
		}

		select {
		case <-time.After(sleep):
		case <-req.Context().Done():
			c.CloseIdleConnections()
			return resp, req.Context().Err()
		}
	}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return resp, err
	}

	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") &&
		!bodyEmpty(resp.Body) {
		if _, ok := resp.Body.(*gzip.Reader); !ok {
			body, err := gzip.NewReader(resp.Body)
			resp.Body.Close()
			resp.Body = body
			return resp, err
		}
	}

	return resp, nil
}

func (c *Client) onAfterResponse(resp *Response, err error) {
	for _, callback := range c.afterResponseCallbacks {
		callback.Exit(resp, err)
	}
}
