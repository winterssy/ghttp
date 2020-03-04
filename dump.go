package ghttp

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/winterssy/bufferpool"
)

var (
	reqWriteExcludeHeaderDump = map[string]bool{
		"Host":              true, // not in Header map anyway
		"Transfer-Encoding": true,
		"Trailer":           true,
	}
)

func dumpRequestLine(req *Request, w io.Writer) {
	fmt.Fprintf(w, "> %s %s %s\r\n", req.Method, req.URL.RequestURI(), req.Proto)
}

func dumpRequestHeaders(req *Request, w io.Writer) {
	host := req.Host
	if req.Host == "" && req.URL != nil {
		host = req.URL.Host
	}
	if host != "" {
		fmt.Fprintf(w, "> Host: %s\r\n", host)
	}

	if len(req.TransferEncoding) > 0 {
		fmt.Fprintf(w, "> Transfer-Encoding: %s\r\n", strings.Join(req.TransferEncoding, ","))
	}
	if req.Close {
		io.WriteString(w, "> Connection: close\r\n")
	}

	for k, vs := range req.Header {
		if !reqWriteExcludeHeaderDump[k] {
			for _, v := range vs {
				fmt.Fprintf(w, "> %s: %s\r\n", k, v)
			}
		}
	}
	io.WriteString(w, ">\r\n")
}

func dumpRequestBody(req *Request, w io.Writer) (err error) {
	buf := bufferpool.Get()
	defer buf.Free()
	err = drainBody(req.Body, buf)
	if err == nil {
		req.SetContent(buf.Bytes())
		_, _ = io.Copy(w, buf)
		io.WriteString(w, "\r\n")
	}
	return
}

func dumpRequest(req *Request, w io.Writer, body bool) (err error) {
	dumpRequestLine(req, w)
	dumpRequestHeaders(req, w)
	if body && !bodyEmpty(req.Body) {
		err = dumpRequestBody(req, w)
	}
	return
}

func dumpResponseLine(resp *Response, w io.Writer) {
	fmt.Fprintf(w, "< %s %s\r\n", resp.Proto, resp.Status)
}

func dumpResponseHeaders(resp *Response, w io.Writer) {
	for k, vs := range resp.Header {
		for _, v := range vs {
			fmt.Fprintf(w, "< %s: %s\r\n", k, v)
		}
	}
	io.WriteString(w, "<\r\n")
}

func dumpResponseBody(resp *Response, w io.Writer) (err error) {
	buf := bufferpool.Get()
	defer buf.Free()
	err = drainBody(resp.Body, buf)
	if err == nil {
		resp.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
		_, _ = io.Copy(w, buf)
		io.WriteString(w, "\r\n")
	}
	return
}

func dumpResponse(resp *Response, w io.Writer, body bool) (err error) {
	dumpResponseLine(resp, w)
	dumpResponseHeaders(resp, w)
	if body && !bodyEmpty(resp.Body) {
		err = dumpResponseBody(resp, w)
	}
	return
}
