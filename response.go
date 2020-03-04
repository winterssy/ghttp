package ghttp

import (
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/winterssy/bufferpool"
	"github.com/winterssy/gjson"
	"golang.org/x/text/encoding"
)

type (
	// Response is a wrapper around an http.Response.
	Response struct {
		*http.Response
		clientTrace *clientTrace
	}
)

// Cookie returns the named cookie provided in resp.
// If multiple cookies match the given name, only one cookie will be returned.
func (resp *Response) Cookie(name string) (*http.Cookie, error) {
	return findCookie(name, resp.Cookies())
}

// Content reads from resp's body until an error or EOF and returns the data it read.
func (resp *Response) Content() ([]byte, error) {
	buf := bufferpool.Get()
	defer buf.Free()
	err := drainBody(resp.Body, buf)
	return buf.Bytes(), err
}

// Text is like Content, but it decodes the data it read to a string given an optional charset encoding.
func (resp *Response) Text(e ...encoding.Encoding) (string, error) {
	b, err := resp.Content()
	if err != nil || len(e) == 0 {
		return b2s(b), err
	}

	b, err = e[0].NewDecoder().Bytes(b)
	return b2s(b), err
}

// JSON decodes resp's body and unmarshals its JSON-encoded data into v.
// v must be a pointer.
func (resp *Response) JSON(v interface{}, opts ...func(dec *gjson.Decoder)) error {
	defer resp.Body.Close()
	return gjson.NewDecoder(resp.Body, opts...).Decode(v)
}

// H is like JSON, but it unmarshals into an H instance.
// It provides a convenient way to read arbitrary JSON.
func (resp *Response) H(opts ...func(dec *gjson.Decoder)) (H, error) {
	var h H
	return h, resp.JSON(&h, opts...)
}

// SaveFile saves resp's body into a file.
func (resp *Response) SaveFile(filename string, perm os.FileMode) (err error) {
	var file *os.File
	file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err == nil {
		defer file.Close()
		err = drainBody(resp.Body, file)
	}
	return
}

// TraceInfo returns the trace info for the request if client trace is enabled.
func (resp *Response) TraceInfo() (traceInfo *TraceInfo) {
	if resp.clientTrace != nil {
		traceInfo = resp.clientTrace.traceInfo()
	}
	return
}

// Dump returns the HTTP/1.x wire representation of resp.
func (resp *Response) Dump(body bool) ([]byte, error) {
	return httputil.DumpResponse(resp.Response, body)
}
